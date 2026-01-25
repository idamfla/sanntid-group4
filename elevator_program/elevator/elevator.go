package elevator

import (
	"fmt"

	// "elevator_program/utilities"
	"elevator_program/elevio"
)

var ascendingButtons = []elevio.ButtonType{elevio.BT_Cab, elevio.BT_HallUp}
var descendingButtons = []elevio.ButtonType{elevio.BT_Cab, elevio.BT_HallDown}

type SortingOrder int

const (
	SO_Ascending SortingOrder = 1
	SO_Descending SortingOrder = -1
)

type ElevatorState int

const (
	ES_Uninitialized ElevatorState = iota
	ES_Idle
	ES_Moving
	ES_DoorOpen
	ES_Obstruction
)

type Elevator struct {
	id            int
	currentFloor  int
	targetFloor   int
	initFloor     int
	floorRequests [][3]bool             // TODO maybe Pending, Running, Completed, NotActive
	lastDirection elevio.MotorDirection // TODO make targetFloor into ButtonEvent
	state         ElevatorState
	emergencyStop bool
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
	e.targetFloor = -1 // TODO maybe make dynamic, variable on init
	e.initFloor = initFloor
	e.floorRequests = make([][3]bool, numFloors)
	e.state = ES_Uninitialized

	e.eventsCh = make(chan ElevatorEvent, 20)

	// e.state = ES_Moving

	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.targetFloor}
}

func (e *Elevator) RunElevatorProgram() {
	go e.ElevatorStateMachine()
	// go e.EventLoop()
	e.StartHardwareEventsListeners()

	done := make(chan struct{})
	<-done
}

// region printing, for debugging
func (s ElevatorState) String() string {
	switch s {
	// case Idle:
	// 		return "idle"
	case ES_Uninitialized:
		return "uninitialized"
	case ES_Idle:
		return "idle"
	case ES_Moving:
		return "moving"
	case ES_DoorOpen:
		return "door open"
	case ES_Obstruction:
		return "obstruction"
	// case ES_EmergencyStop:
	// 	return "emergency stop"
	default:
		return "unknown"
	}
}

func (e Elevator) String() string {
	s := fmt.Sprintf(
		`Elevator
	id: %d
	current floor: %d
	target floor: %d
	init floor: %d
	last moving dir: %s
	state: %s
`,
		e.id, e.currentFloor, e.targetFloor, e.initFloor, e.lastDirection, e.state)

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