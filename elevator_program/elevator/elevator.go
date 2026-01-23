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
	lastMovingDir elevio.MotorDirection // TODO make targetFloor into ButtonEvent
	state         ElevatorState

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

func (e *Elevator) RunElevatorProgram(port string, id int, numFloors int, initFloor int) {
	// numFloors := 4

	// "localhost:15657"
	elevio.Init("localhost:"+port, numFloors)

	e.InitElevator(id, numFloors, initFloor)
	go e.ElevatorStateMachine()

	go func() {
		drv_buttons := make(chan elevio.ButtonEvent)
		go elevio.PollButtons(drv_buttons)
		for btn := range drv_buttons {
			e.eventsCh <- ElevatorEvent{Type: EV_ButtonPress, Floor: btn.Floor, Button: btn.Button}
		}
	}()

	go func() {
		drv_floors := make(chan int)
		go elevio.PollFloorSensor(drv_floors)
		for f := range drv_floors {
			e.eventsCh <- ElevatorEvent{Type: EV_FloorSensor, Floor: f}
		}
	}()

	go func() {
		drv_obstr := make(chan bool)
		go elevio.PollObstructionSwitch(drv_obstr)
		for obstr := range drv_obstr {
			e.eventsCh <- ElevatorEvent{Type: EV_Obstruction, Obstruction: obstr}
		}
	}()

	go func() {
		drv_stop := make(chan bool)
		go elevio.PollStopButton(drv_stop)
		for s := range drv_stop {
			e.eventsCh <- ElevatorEvent{Type: EV_EmergencyStop, EmergencyStop: s}
		}
	}()

	select {}
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
		e.id, e.currentFloor, e.targetFloor, e.initFloor, e.lastMovingDir, e.state)

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
