package elevator

import (
	"fmt"

	// "elevator_program/utilities"
	"elevator_program/elevio"
)

type Elevator struct {
	id            int
	currentFloor  int
	targetFloor   int
	initFloor     int
	floorRequests [][3]bool             // TODO maybe Pending, Running, Completed, NotActive
	lastDirection elevio.MotorDirection // TODO make targetFloor into ButtonEvent
	state         ElevatorState
	emergencyStop bool // TODO add InBetweenFloors bool, also make sure the order of all is good and that funcitons make sense, name etc

	eventsCh chan ElevatorEvent
	stateMachineCh chan ElevatorEvent

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
	e.stateMachineCh = make(chan ElevatorEvent, 20)

	// e.state = ES_Moving

	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.targetFloor}
}

func (e *Elevator) RunElevatorProgram() {
	go e.ElevatorStateMachine()
	go e.EventLoop()
	e.StartHardwareEventsListeners()

	done := make(chan struct{})
	<-done
}

// region print elevator, for debugging
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