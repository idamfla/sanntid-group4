package elevator

import (
	"fmt"
	"time"

	"elevator_program/elevio"
	"elevator_program/utilities"
)

type ButtonStatus int

const (
	NotActive ButtonStatus = iota
	Pending
	Running
)

// Uses to add a new elevator to our system
type System struct {
	hallRequests [][2]ButtonStatus
	Elevators    map[int]ElevatorsStatus
	// mutex        sync.Mutex // Add a mutex to protect shared data
}

type Elevator struct {
	id int
	// TODO temp need to know the ip using the id
	ipToId map[string]int

	inBetweenFloors bool
	currentFloor    int
	nextTarget      elevio.ButtonEvent
	direction       elevio.MotorDirection
	initFloor       int

	hallRequests [][2]ButtonStatus // TODO Should remove these
	cabRequests  []ButtonStatus    // TODO Should remove these

	doorState DoorState
	doorTimer time.Time

	elevatorState    ElevatorState
	obstruction      bool
	emergencyStop    bool // TODO fade out ... just figure out how to set state to ES_EmergencyStop, unset it
	hardwareEventsCh chan HardwareEvent

	msgRecieveCh chan utilities.Message
	msgSendCh    chan utilities.Message

	isMaster          bool
	connectedToMaster bool
	elevatorRegistry  map[int]ElevatorsStatus // TODO Was string, could also make it uint

	// TODO Trying to split ut the code
	protocol *Protocol // TODO should we remove this one from elevator struct and put in a different package
	system   System
}

func (e *Elevator) InitElevator(id int, numFloors int, initFloor int) {
	e.id = id
	e.currentFloor = -1
	e.nextTarget = elevio.ButtonEvent{Floor: -1}
	e.initFloor = initFloor
	e.doorTimer = time.Time{}
	e.hallRequests = make([][2]ButtonStatus, numFloors)
	e.cabRequests = make([]ButtonStatus, numFloors)

	e.system.hallRequests = make([][2]ButtonStatus, numFloors)
	e.system.Elevators = make(map[int]ElevatorsStatus)
	e.system.Elevators[id] = ElevatorsStatus{
		CabRequests: make([]ButtonStatus, numFloors),
		Id:          id,
	}
	e.isMaster = false

	e.protocol = &Protocol{
		ackArray: make(map[int]int),
	} // Initialize the Protocol field

	// e.elevatorState = ES_Uninitialized

	e.hardwareEventsCh = make(chan HardwareEvent, 20)

	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.nextTarget}

	e.clearAllLamps(elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab)
}

func (e *Elevator) RunElevatorProgram() {
	fmt.Println("RUNNING ELEVATOR PROGRAM")
	go e.RunHardwareEventLoop()
	go e.RunDoorStateMachine()
	go e.RunElevatorStateMachine()
	e.StartHardwareEventsListeners()
	time.Sleep(10 * time.Second)

	// Temp for testing msgHandler
	// go e.TestMsgHandler(4)
	go e.TestMsgHandler_Master(4)
	done := make(chan struct{})
	<-done
}

// Temp for printing ButtonStatus
func (r ButtonStatus) String() string {
	switch r {
	case NotActive:
		return "NotActive"
	case Pending:
		return "Pending"
	case Running:
		return "Running"
	default:
		return "Unknown"
	}
}

// region printing, for debugging
func (e Elevator) String() string {
	s := fmt.Sprintf(
		`Elevator
	id: %d
	in between floors: %t
	current floor: %d
	target: %d, %s
	init floor: %d
	last moving dir: %s
	door state: %s
	elevator state: %s
`,
		e.id, e.inBetweenFloors, e.currentFloor, e.nextTarget.Floor, e.nextTarget.Button, e.initFloor, e.direction, e.doorState, e.elevatorState)

	for f := 0; f < len(e.hallRequests); f++ {
		req := e.system.hallRequests[f]
		cab := e.system.Elevators[e.id].CabRequests[f]

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
