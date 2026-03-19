package elevator

import (
	"fmt"
	"sync"
	"time"

	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/system"
	"elevator_program/types"
)

type Elevator struct {
	Id string
	Ip string // TODO Think we should have this one here
	// TODO temp need to know the ip using the id
	IpRegistery map[string]string

	offline         bool
	scheduleRestart bool

	inBetweenFloors bool
	currentFloor    int
	// nextTarget      elevio.ButtonEvent
	// direction elevio.MotorDirection
	initFloor  int
	nextTarget elevio.ButtonEvent
	direction  elevio.MotorDirection
	numFloors  int

	hallRequests [][2]types.ButtonStatus // TODO Should remove these
	cabRequests  []types.ButtonStatus    // TODO Should remove these

	doorState DoorState
	doorTimer time.Time

	//temp Need to time how long you have lost communiction
	lostComsTimer      time.Time
	ackCounterLostComs int

	// TODO Maybe temp need to notify protocol to send something
	SendToCoordinator chan message.ElevatorMessage

	// elevatorState    types.ElevatorState
	obstruction              bool
	emergencyStop            bool // TODO fade out ... just figure out how to set state to ES_EmergencyStop, unset it
	hardwareEventsCh         chan HardwareEvent
	hardwareListenersStarted bool

	FaultMsg chan message.FaultMessage

	// MsgRecieveCh chan message.ElevatorMessage
	// msgSendCh    chan message.ElevatorMessage
	// MsgRecieveCh chan session.ElevatorPacket // Update the channel type, wait should this one be IncomingPacket, do i need to debug and encode this one?

	IsMaster          bool
	connectedToMaster bool
	IsOnline          bool

	System          system.System
	currentMasterID string

	stop      chan struct{}
	runningMu sync.Mutex
	isRunning bool

	wg sync.WaitGroup

	recoveryCfg RecoveryConfig

	restartReason         RestartReason
	softRestartInProgress bool
	softRestartAttempts   int

	recoveryAwaitingProof bool
	recoveryVerified      bool
	lastRecoveryAttempt   time.Time
	recoveryMu            sync.Mutex
	// Server *server.Server // TODO be carefull with pass by value functions, locks
}

func (e *Elevator) InitElevator(id string, numFloors int, initFloor int, ip string, port int) { // TODO Changed port to string, hope everything works
	e.Id = id
	e.Ip = ip
	e.currentFloor = -1
	// e.nextTarget = elevio.ButtonEvent{Floor: -1}
	e.initFloor = initFloor
	e.doorTimer = time.Time{}
	e.hallRequests = make([][2]types.ButtonStatus, numFloors)
	e.cabRequests = make([]types.ButtonStatus, numFloors)
	e.numFloors = numFloors

	e.IsMaster = false

	e.System.InitSystem(id, "192.168.0.1", numFloors)

	// e.elevatorState = types.ES_Uninitialized

	//Temp init door timer
	e.lostComsTimer = time.Time{}
	e.ackCounterLostComs = 0

	e.IpRegistery = make(map[string]string)

	e.SendToCoordinator = make(chan message.ElevatorMessage, 10)
	e.FaultMsg = make(chan message.FaultMessage, 20)

	e.hardwareEventsCh = make(chan HardwareEvent, 20)

	e.recoveryCfg = DefaultRecoveryConfig
	e.restartReason = RestartReasonNone
	e.IsOnline = false
	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.MsgRecieveCh = make(chan session.ElevatorPacket, 10) // Match the expected type

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.nextTarget}
	// e.MsgRecieveCh = make(chan message.ElevatorMessage, 10) // TODO Should have this in the code

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
	go e.fault_loop()

	e.isRunning = true
}

func (e *Elevator) resetRuntimeState(numFloors int) {
	e.offline = false
	e.scheduleRestart = false

	e.inBetweenFloors = false
	e.currentFloor = elevio.GetFloor()
	//e.nextTarget = elevio.ButtonEvent{Floor: -1, Button: elevio.BT_Cab}
	//e.direction = elevio.MD_Stop

	e.doorState = DS_Closed
	e.doorTimer = time.Time{}

	e.lostComsTimer = time.Time{}
	e.ackCounterLostComs = 0

	e.obstruction = false
	e.emergencyStop = false

	e.IsMaster = false
	e.connectedToMaster = false
	e.IsOnline = false
	e.currentMasterID = ""

	//e.hallRequests = make([][2]types.ButtonStatus, numFloors)
	//e.cabRequests = make([]types.ButtonStatus, numFloors)

	//e.System = system.System{}
	//e.System.InitSystem(e.Id, e.Ip, numFloors)

	e.IpRegistery = make(map[string]string)

	//e.clearAllLamps(elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab)
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	elevio.SetMotorDirection(elevio.MD_Stop)

	if e.currentFloor >= 0 {
		elevio.SetFloorIndicator(e.currentFloor)
	}

	e.SendToCoordinator = make(chan message.ElevatorMessage, 10)

}

func (e *Elevator) stopRuntimeLoops() {
	e.runningMu.Lock()
	defer e.runningMu.Unlock()

	if !e.isRunning {
		return
	}

	close(e.stop)
	e.wg.Wait()

	e.isRunning = false
}

// region printing, for debugging
func (e *Elevator) String() string {
	e.System.Mutex.RLock()
	defer e.System.Mutex.RUnlock()

	elevStatus := e.System.Elevators[e.Id]
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
		// e.Id, e.inBetweenFloors, e.currentFloor, e.System.Elevators[e.Id].Target.Floor, e.System.Elevators[e.Id].Target.Button, e.initFloor, e.System.Elevators[e.Id].Direction, e.doorState, e.System.Elevators[e.Id].State)
		e.Id, e.inBetweenFloors, e.currentFloor, elevStatus.Target.Floor, elevStatus.Target.Button, e.initFloor, elevStatus.Direction, e.doorState, elevStatus.State)
	for f := 0; f < len(e.hallRequests); f++ {
		req := e.System.HallRequests[f]
		// cab := e.System.Elevators[e.Id].CabRequests[f]
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

// endregion
