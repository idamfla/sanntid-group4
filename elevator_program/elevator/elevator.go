package elevator

import (
	"fmt"
	"time"

	// "elevator_program/utilities"
	"elevator_program/elevio"
)

// ------------------------
// Sorting
// ------------------------
var ascendingButtons = []elevio.ButtonType{elevio.BT_Cab, elevio.BT_HallUp}
var descendingButtons = []elevio.ButtonType{elevio.BT_Cab, elevio.BT_HallDown}

type SortingOrder int

const (
	SO_Ascending  SortingOrder = 1
	SO_Descending SortingOrder = -1
)

type Elevator struct {
	id int

	inBetweenFloors bool
	currentFloor    int
	nextTarget      elevio.ButtonEvent // TODO maybe a targetRequest, of request{Floor: f, MotorDirection: md}
	initFloor       int
	lastDirection   elevio.MotorDirection // TODO make nextTarget into ButtonEvent

	startTime time.Time

	floorRequests [][3]bool // TODO maybe Pending, Running, Completed, NotActive

	doorState     DoorState
	state         ElevatorState
	obstruction   bool
	emergencyStop bool // TODO fade out ... just figure out how to set state to ES_EmergencyStop, unset it
	/*
		TODO add InBetweenFloors bool,
			also make sure the order of all is good and that funcitons make sense, name etc
		always update current floor, but maybe also have a lastValidFloor or something
	*/
	eventsCh chan ElevatorEvent

	// StatusChan chan utilities.StatusMsg
	// TaskChan chan utilities.TaskMsg
}

func (e *Elevator) InitElevator(id int, numFloors int, initFloor int) {
	e.id = id
	e.currentFloor = -1
	e.nextTarget = elevio.ButtonEvent{Floor: -1}
	e.initFloor = initFloor
	e.startTime = time.Time{}
	e.floorRequests = make([][3]bool, numFloors)
	// e.state = ES_Uninitialized

	e.eventsCh = make(chan ElevatorEvent, 20)

	// e.state = ES_Moving

	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.nextTarget}

	e.clearAllLamps(elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab)
}

func (e *Elevator) RunElevatorProgram() {
	fmt.Println("RUNNING ELEVATOR PROGRAM")
	go e.RunEventLoop()
	go e.RunDoorStateMachine()
	go e.RunElevatorStateMachine()
	e.StartHardwareEventsListeners()

	done := make(chan struct{})
	<-done
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
	state: %s
`,
		e.id, e.inBetweenFloors, e.currentFloor, e.nextTarget.Floor, e.nextTarget.Button, e.initFloor, e.lastDirection, e.doorState, e.state)

	for f, req := range e.floorRequests {
		s += fmt.Sprintf(
			"	floor %d: [Up:%t Down:%t Cab:%t]\n",
			f,
			req[elevio.BT_HallUp],
			req[elevio.BT_HallDown],
			req[elevio.BT_Cab],
		)
	}

	return s
}

// endregion
