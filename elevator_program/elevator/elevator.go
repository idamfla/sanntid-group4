package elevator

import (
	"fmt"
	"sync"
	"time"

	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/system"
)

type Elevator struct {
	Id string
	Ip string
	// IpRegistery map[string]string

	scheduleRestart bool

	inBetweenFloors bool
	currentFloor    int
	initFloor       int
	NumFloors       int

	direction elevio.MotorDirection

	doorState DoorState
	doorTimer time.Time

	mu sync.Mutex

	SendToCoordinator chan message.ElevatorMessage

	obstruction              bool
	emergencyStop            bool
	hardwareEventsCh         chan HardwareEvent
	hardwareListenersStarted bool

	FaultMsg chan message.FaultMessage

	IsMaster          bool
	connectedToMaster bool // TODO i am not sure what we actually do with this one?
	IsOnline          bool

	System          system.System
	currentMasterID string

	stop      chan struct{}
	runningMu sync.Mutex
	isRunning bool

	wg sync.WaitGroup

	motorWatchdogFloor int
	motorWatchdogTimer time.Time

	// recoveryCfg RecoveryConfig

	// restartReason         RestartReason
	// softRestartInProgress bool
	// softRestartAttempts   int

	// recoveryAwaitingProof bool
	// recoveryVerified      bool
	// lastRecoveryAttempt   time.Time
	// recoveryMu            sync.Mutex
}

func (e *Elevator) InitElevator(id string, numFloors int, initFloor int, ip string, port int) {
	e.Id = id
	e.Ip = ip

	e.currentFloor = -1
	e.initFloor = initFloor
	e.doorTimer = time.Time{}
	e.NumFloors = numFloors

	e.IsMaster = false
	e.connectedToMaster = false

	e.System.InitSystem(id, ip, numFloors)

	// e.IpRegistery = make(map[string]string)

	e.SendToCoordinator = make(chan message.ElevatorMessage, 10)
	e.FaultMsg = make(chan message.FaultMessage, 20)

	e.hardwareEventsCh = make(chan HardwareEvent, 20)

	// e.recoveryCfg = DefaultRecoveryConfig
	// e.restartReason = RestartReasonNone
	e.IsOnline = false

	e.clearAllLamps(elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab)

}

func (e *Elevator) RunElevatorProgram() {

	e.runningMu.Lock()
	defer e.runningMu.Unlock()

	if e.isRunning {
		return
	}

	fmt.Println("RUNNING ELEVATOR PROGRAM")

	e.stop = make(chan struct{})
	if !e.hardwareListenersStarted {
		e.StartHardwareEventsListeners()
		e.hardwareListenersStarted = true
	}

	e.wg.Add(4)
	go e.RunHardwareEventLoop()
	go e.RunDoorStateMachine()
	go e.RunElevatorStateMachine()
	// go e.fault_loop()

	e.isRunning = true
}

func (e *Elevator) resetRuntimeState(numFloors int) {
	e.scheduleRestart = false

	e.inBetweenFloors = false
	e.currentFloor = elevio.GetFloor()
	e.direction = elevio.MD_Stop

	e.doorState = DS_Closed
	e.doorTimer = time.Time{}

	e.obstruction = false
	e.emergencyStop = false

	e.IsMaster = false
	e.connectedToMaster = false
	e.IsOnline = false
	e.currentMasterID = ""
	// e.IpRegistery = make(map[string]string)

	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	elevio.SetMotorDirection(elevio.MD_Stop)

	if e.currentFloor >= 0 {
		elevio.SetFloorIndicator(e.currentFloor)
	}

	for {
		select {
		case <-e.SendToCoordinator:
		default:
			goto drained
		}
	}
drained:
}

func (e *Elevator) stopRuntimeLoops() {
	e.runningMu.Lock()

	if !e.isRunning {
		e.runningMu.Unlock()
		return
	}

	e.isRunning = false
}

func (e *Elevator) String() string {
	e.System.Mutex.RLock()
	defer e.System.Mutex.RUnlock()

	elevStatus := e.System.Elevators[e.Ip]
	s := fmt.Sprintf(
		`Elevator
	id: %s
	in between floors: %t
	current floor: %d
	target: %d, %s
	init floor: %d
	last moving dir: %s
	door state: %s
	elevator state: %s
`,
		e.Id, e.inBetweenFloors, e.currentFloor, elevStatus.Target.Floor, elevStatus.Target.Button, e.initFloor, elevStatus.Direction, e.doorState, elevStatus.State)
	for f := 0; f < e.NumFloors; f++ {
		req := e.System.HallRequests[f]
		cab := elevStatus.CabRequests[f]

		s += fmt.Sprintf(
			"	floor %d: [Up:%s Down:%s Cab:%s]\n",
			f,
			req[elevio.BT_HallUp],
			req[elevio.BT_HallDown],
			cab,
		)
	}

	return s
}
