# Software Design Document - Elevator Program

## High-level architecture

The system is layered like this:

```
┌─────────────────────────────────────────────┐
│                 main / config               │
│          Parse flags, spawn processes       │
├─────────────────────────────────────────────┤
│               Coordinator                   │
│    Connects between elevator and network    │
├──────────────────────┬──────────────────────┤
│      Elevator        │     UDP layer        │
│  State machines,     │  Server, sessions,   │
│  hardware, faults    │  packets, peers      │
├──────────────────────┴──────────────────────┤
│           System  (shared state)            │
├─────────────────────────────────────────────┤
│         elevio  (hardware driver)           │
└─────────────────────────────────────────────┘
```

**Elevator** owns the physical control: motors, doors, buttons, and sensors.
**Coordinator** routes messages between the elevator logic and the network.
**System** is the shared world-view: who is where, which buttons are pressed, who is doing what.

---

## Project layout

```
elevator_program/
├── main.go                        Entry point
├── config
│   └── config.go                  CLI flags, child-process spawning
│
├── coordinator/
│   ├── coordinator.go             Init, start, close
│   ├── message_handler.go         Master/slave message dispatch
│   ├── send_message.go            Outgoing message routing
│   └── task_monitor.go            Per-task timeout watchdog
│
├── elevator/
│   ├── elevator.go                Core struct, init, lifecycle
│   ├── state_machine.go           Online & offline state machines
│   ├── door.go                    Door open/close/obstruction FSM
│   ├── hardware_event.go          Button press, floor sensor, e-stop
│   ├── hardware.go                Polls hardware and emits events
│   ├── scan.go                    SCAN algorithm for next target
│   ├── target.go                  "Who should take this task?" logic
│   ├── clear_request.go           Mark requests as done
│   ├── fault_loop.go              Fault dispatcher goroutine
│   ├── fault_actions.go           Recovery: soft restart, hard restart
│   ├── protocol_callbacks.go      Network event hooks
│   ├── master_logic.go            Periodic master broadcasts
│   └── clear_lamp.go              Lamp control helpers
│
├── system/
│   ├── system.go                  System struct, Snapshot(), init
│   └── update_System.go           Mutating helpers (set status, assign, register)
│
├── message/
│   ├── elevator_message.go        ElevatorMessage + message types enum
│   └── fault_message.go           FaultMessage + fault types enum
│
├── types/
│   ├── types.go                   ElevatorState, ButtonStatus enums
│   └── elevator_status.go         ElevatorsStatus (per-elevator snapshot)
│
├── elevio/                     
│    ├── elevator_io.go            Hardware I/O (motor, lamps, sensors)                  
│    └── elevio_type_converter.go  Type conversion between driver and system
│
├── udp/
│   ├── server/                    UDP server, session lifecycle, peer tracking
│   ├── session/                   Point-to-point & broadcast sessions
│   ├── packet/                    Encode/decode, packet type enum
│   ├── peerinfo/                  Peer metadata
│   └── timer/                     Reusable timer wrapper
│
├── fault/
│   └── restart.go                 Hard restart (re-exec the binary)
│
├──  utilities/
│    └── utilities.go              Small helpers (Abs, etc.)

```

---

## How it starts up

### Running the program

```bash
# Option 1 -- Launcher mode (spawns N simulator elevators in separate terminals)
go run . --n 3 --baseport 15657 --floors 4 --initfloor 3

# Option 2 -- Single elevator (for example started by the launcher)
go run . --id 1 --ip localhost --port 15657 --floors 4 --initfloor 3

# Option 3 -- Explicit addresses for physical elevators
go run . --addrs "192.168.0.1:15657,192.168.0.2:15657"
```

When `--id` is 0 (the default), the process acts as a **launcher**: it spawns `--n` child processes, each in its own terminal window. When `--id` is non-zero, the process runs a single elevator.

### Boot sequence

The order matters, each step depends on the previous one.

```
1.  elevio.Init(addr, numFloors)
    Connect to elevator hardware (TCP socket to simulator or real controller).

2.  elevator.InitElevator(id, numFloors, initFloor, ip, port)
    Create channels, initialize System, clear all lamps.

3.  coordinator.InitCoordinator()
    Create message channels, port registry, TaskMonitor (15 s timeout).

4.  coordinator.StartServer(ip, port, id)
    Open two UDP sockets: one for unicast (ip:port), one for broadcast (:3000).

5.  coordinator.Start(&elevator)
    Launch three goroutines:
      - MessageListener   reads from network → handles messages
      - sendListener      reads from elevator → sends over network
      - Server.Start      the UDP server main loop

6.  elevator.RunElevatorProgram()
    Launch four goroutines:
      - RunHardwareEventLoop      buttons, sensors, e-stop
      - RunDoorStateMachine       door FSM (50 ms tick)
      - RunElevatorStateMachine   motor control FSM (50 ms tick)
      - fault_loop                listens for fault messages
```

Once the state machine starts, the elevator drives to `initFloor`, transitions to `Idle`, and asks the master for work.

---

## Master / slave model

One elevator is the **master**. It owns the authoritative copy of the system state, decides who handles which hall call, and broadcasts updates. Everyone else is a **slave**: they report button presses and status changes, execute assigned tasks, and request new work when idle.

### How a master is elected

When an elevator starts (or loses its master), it broadcasts a `WhoIsMaster` packet. If no one answers within 500 ms, an election happens: all peers that responded are compared, and the one with the **lowest network address** wins. The winner announces itself with `IAmMaster`, and peers ACK.

---

## The elevator state machine

Each elevator runs one of two state machines depending on whether it has network connectivity.

### Online mode

The elevator trusts the master for task assignments.

```
 ┌──────────────────┐
 │ ES_Uninitialized │  Drive to initFloor
 └────────┬─────────┘
          │  arrived
          v
 ┌──────────────────┐
 │     ES_Idle      │◄──────────────┐
 │  waiting for     │               │
 │  target from     │               │  arrived at target
 │  master          │               │  → open door
 └────────┬─────────┘               │  → finishedTask()
          │  target assigned        │  → ask for new task
          v                         │
 ┌──────────────────┐               │
 │    ES_Moving     │───────────────┘
 │  driving towards │
 │  target floor    │
 └──────────────────┘
```

When the elevator reaches its target it calls `finishedTask()`, which:
1. Sends `ButtonPress(NotActive)` -- "I'm done with this task"
2. Clears the target locally
3. Sends `TaskRequest` -- "give me something new"

### Offline mode

If the elevator loses the network, it switches to offline mode and only handles **cab requests** (no hall calls). It scans its own cab request list, picks the closest one, drives there, and repeats. Once all cab orders are done and the door is closed, it tries to reconnect by restarting.
### Door FSM

The door has its own state machine running on a 50 ms tick:

```
DS_Closed ↔ DS_Opening → DS_Open → DS_Closing → DS_Closed
                            ↓
                       DS_Obstruction  (while sensor active)
```

The door stays open for a fixed timer. If the obstruction sensor fires while the door is open, it waits until the obstruction clears before closing.

---

## Message flow

### Message types at a glance

| Message | What it means | Direction |
|---|---|---|
| `StatusReport` | "Here's my current floor/state/direction" | Both ways |
| `ButtonPress` | "Someone pressed a button" | Slave → Master |
| `TaskUpdate` | "This task is now assigned / completed / timed out" | Master → Slaves |
| `TaskRequest` | "I'm free, give me work" | Slave → Master |
| `LostComs` | "I can't reach someone" | Broadcast |
| `ElevatorLost` | "This elevator disappeared" | Broadcast |
| `NewToChannel` | "I just joined, sync me up" | Broadcast |

### What the master does with each message

**ButtonPress**: A slave noticed a new button press.
- If it's a cab button: check whether it's on the requester's current path (`IsNewTargetBetterCab`). Mark it `Running` or `Pending`.
- If it's a hall button: run `ClosestToTarget` over all elevators, assign the winner, broadcast the assignment as a `TaskUpdate`.

**TaskRequest**: An elevator finished its task and wants more work.
- Take a snapshot of the current hall requests.
- Run `GetNextTargetFloor` for the requesting elevator.
- If a task is found, broadcast it as a `TaskUpdate(Running)`.

**TaskUpdate**: Comes back through the system (master processes its own broadcasts too).
- `Running`: start a 15-second watchdog timer via `TaskMonitor`.
- `NotActive`: cancel the timer, the task is done.
- If the timer fires before the task completes, the task is reset to `Pending` so someone else can pick it up.

**StatusReport**: Just update the system state and re-broadcast to slaves.

**NewToChannel**: A new elevator appeared. Register it, send a full system snapshot back.

### What a slave does with each message

**TaskUpdate**: If `Running` and addressed to me: set it as my target (`SetRequestAsTarget`). Otherwise just update the local system copy.

**StatusReport**: Update local system state.

**NewToChannel**: If it's about me: initialize my state from the master's snapshot. Otherwise just note the new peer.

---

## Task assignment logic

### The SCAN algorithm (`scan.go`)

When the master needs to find the next task for an elevator, it uses `GetNextTargetFloor`:

- **Elevator is idle**: pick the closest floor with a `Pending` request (cab or hall).
- **Elevator is moving up**: scan current floor → continue up → sweep down → sweep back up. This is the classic disk-scheduling SCAN algorithm.
- **Elevator is moving down**: mirror of the above.

The scan skips hall requests marked `Running` that belong to another elevator (anti-stealing).

### Choosing which elevator gets a hall call (`target.go`)

`ClosestToTarget` loops over all elevators and asks: "can this elevator take the new task, and how far away is it?"

- Idle elevators are always preferred over moving ones.
- A moving elevator can only take a task if it's **on the way** (between current position and current target, in the right direction).
- The elevator with the shortest distance wins.

---

## Network layer (UDP)

### Packet types

The protocol uses these packet types:

| Packet | Purpose |
|---|---|
| `WhoIsMaster` | Broadcast discovery |
| `IAmMaster` / `MasterAck` | Master announcement handshake |
| `BroadcastUpdate` / `BroadcastAck` / `BroadcastCommit` / `BroadcastDone` | 2-phase state broadcast |
| `SlaveUpdate` | Slave → master point-to-point |
| `RequestTaskExecution` | Slave → master task request |

### 2-phase broadcast

The master doesn't just fire-and-forget state updates. It uses a simple 2-phase protocol:

```
  Master                          Slaves
    │                               │
    ├── BroadcastUpdate ───────────→│
    │                               │
    │←─────── BroadcastAck ─────────┤   wait for all peers to ACK
    │                               │
    ├── BroadcastCommit ───────────→│   "ok, apply this update"
    │                               │
    │←─────── BroadcastDone ────────┤   "done"
```

This ensures all elevators see the same state before anyone acts on it.

### Timeouts

| Name | Value | What happens when it fires |
|---|---|---|
| Packet retry | 500 ms, max 10 tries | Resend the packet |
| Broadcast ACK | 2 s | Peer considered unresponsive |
| Broadcast commit | 3 s | Commit phase aborted |
| Master election | 500 ms | Start election |
| Task watchdog | 15 s | Reset task to Pending |

---

## Fault tolerance

### What can go wrong and what happens

| Fault | How it's detected | What the system does |
|---|---|---|
| Network loss | UDP timeouts | Elevator goes offline, finishes cab orders, then tries to recover |
| Master gone | No response to messages | `connectedToMaster = false`, triggers new election |
| Elevator died | Broadcast ACK missing | Remove peer from system, reassign tasks later |
| Task stuck | 15 s watchdog fires | Reset task to `Pending` so another elevator picks it up |
| Motor failure | Task timeout / no floor sensor | Stop motor, go offline, schedule restart |

### Recovery: soft restart → hard restart

1. **Soft restart**: stop all goroutines, reset state, re-launch `RunElevatorProgram()`. Then wait 3 seconds for proof that the elevator actually works (it must transition to `Moving`).
2. If proof arrives: success, reset the attempt counter.
3. If no proof after 3 seconds, or max attempts reached (1 by default): **hard restart** -- the process re-executes itself (`os.Exec`).

---

## Shared state and synchronization

All shared state lives in the `System` struct, protected by a `sync.RWMutex`.

```go
type System struct {
    HallRequests [][2]ButtonStatus           // [floor][up/down]
    Elevators    map[string]ElevatorsStatus   // elevatorID → status
    Mutex        sync.RWMutex
}
```

The key mutation methods:

- **`SetStatusReport(id, status)`** -- update one elevator's status (floor, direction, state).
- **`SetRequestStatus(id, status, task)`** -- mark a button as NotActive / Pending / Running.
- **`SetRequestAsTarget(id, task)`** -- assign a task to an elevator. If it already had a target, that one goes back to Pending.
- **`RegisterAndSyncElevator(msg)`** -- a new elevator joined; add it to the map and return a full snapshot.
- **`Snapshot()`** -- deep-copy the entire system for safe reading without holding the lock.

---

## Goroutines and channels

Each elevator instance runs **7 goroutines** (4 elevator + 3 coordinator):

```
 Hardware pollers (buttons, floor sensor, obstruction, stop)
        │
        v
 hardwareEventsCh ──→ RunHardwareEventLoop ──┐
                                              │
 (50ms tick) ────────→ RunElevatorStateMachine ──→ SendToCoordinator
                                              │         │
 (50ms tick) ────────→ RunDoorStateMachine    │         v
                                              │    sendListener
 FaultMsg ───────────→ fault_loop             │     │         │
                                              │  sendAsMaster  sendAsSlave
                                              │     │              │
                                              │     v              v
                                              │       UDP network
                                              │           │
                                              │           v
                                              │     msgRecieveCh
                                              │           │
                                              │     MessageListener
                                              │           │
                                              │     MessageHandler
                                              │     ├─ handleAsMaster
                                              │     └─ handleAsSlave
                                              │           │
                                              └───────────┘
                                             (updates System, may send replies)
```

### Channel summary

| Channel | Buffer | From → To | Carries |
|---|---|---|---|
| `hardwareEventsCh` | 20 | HW pollers → event loop | Button presses, floor arrivals, e-stop |
| `SendToCoordinator` | 10 | Elevator logic → coordinator | Outgoing ElevatorMessages |
| `FaultMsg` | 20 | Network / watchdogs → fault loop | Fault events |
| `msgRecieveCh` | 10 | UDP server → message listener | Incoming ElevatorPackets |
| `stop` | 0 | Lifecycle → all goroutines | Shutdown signal (close to broadcast) |
