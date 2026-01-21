package elevator

import (
	"fmt"

	// "elevator_program/utilities"	
	"elevator_program/elevio"
)

type Elevator struct {
	id int
	currentFloor int
	targetFloor  int
	initFloor int
	floorRequests [][3]bool
	lastMovingDir elevio.MotorDirection
	state    ElevatorState

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

func (e *Elevator) ClearFloor(f int) {
	e.floorRequests[f][elevio.BT_Cab] = false

	if f == 0 {
		e.floorRequests[f][elevio.BT_HallUp] = false
	} else if f == len(e.floorRequests) - 1 { 
		e.floorRequests[f][elevio.BT_HallDown] = false
	} else {
		switch e.lastMovingDir {
		case elevio.MD_Up:
			e.floorRequests[f][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			e.floorRequests[f][elevio.BT_HallDown] = false
		}
	}
}

func (e *Elevator) ClearButtonLamp(f int) {
	elevio.SetButtonLamp(elevio.BT_Cab, e.currentFloor, false)

	if f == 0 {
		elevio.SetButtonLamp(elevio.BT_HallUp, f, false)
	} else if f == len(e.floorRequests) - 1 {
		elevio.SetButtonLamp(elevio.BT_HallDown, f, false)
	} else {
		switch e.lastMovingDir {
		case elevio.MD_Up:
			elevio.SetButtonLamp(elevio.BT_HallUp, e.currentFloor, false)
		case elevio.MD_Down:
			elevio.SetButtonLamp(elevio.BT_HallDown, e.currentFloor, false)
		}
	}
}

func (e *Elevator) UpdateLastDirection(nextTarget elevio.ButtonEvent, dir elevio.MotorDirection) {
	if nextTarget.Floor == -1 { return }

	if dir != elevio.MD_Stop {
		e.lastMovingDir = dir
	}
	
	if nextTarget.Floor != e.currentFloor { return }
	switch nextTarget.Button {
	case elevio.BT_HallUp:
		e.lastMovingDir = elevio.MD_Up
	case elevio.BT_HallDown:
		e.lastMovingDir = elevio.MD_Down
	}
}

func (e *Elevator) UpdateTargetFloor() elevio.ButtonEvent {
    nextTarget := e.GetNextTargetFloor()
    if nextTarget.Floor != -1 {
        e.targetFloor = nextTarget.Floor
    }
    return nextTarget
}

func (e *Elevator) ElevatorStateMachine() {
	for {
		var ev ElevatorEvent
		select {
		case ev = <-e.eventsCh:
			switch ev.Type {
			case EV_FloorSensor:
				if ev.Floor != -1 { e.currentFloor = ev.Floor }
			case EV_ButtonPress:
				e.floorRequests[ev.Floor][ev.Button] = true
				elevio.SetButtonLamp(ev.Button, ev.Floor, true)
			case EV_Obstruction:
				if ev.Obstruction {
					e.state = ES_Obstruction
				}
			case EV_EmergencyStop:
				if ev.EmergencyStop {
					elevio.SetStopLamp(true)
					e.state = ES_EmergencyStop
					continue
				}
			case EV_TaskAssigned:
				continue
			case EV_TaskCompleted:
				continue
			}
		default:
		}

		switch e.state {
		case ES_Uninitialized:
			if e.currentFloor == -1 {
				elevio.SetMotorDirection(elevio.MD_Down) // always move down until a floor sensor triggers
				continue
			}
			
			if e.currentFloor < e.initFloor {
				elevio.SetMotorDirection(elevio.MD_Up)
			} else if e.currentFloor > e.initFloor {
				elevio.SetMotorDirection(elevio.MD_Down)
			} else {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.ClearFloor(e.currentFloor)
				e.ClearButtonLamp(e.currentFloor)

				e.state = ES_Idle
			}

		case ES_Idle:
			nextTarget := e.UpdateTargetFloor() // sets targetFloor
			dir := e.GetMotion()
			if dir != elevio.MD_Stop {
				e.state = ES_Moving
				elevio.SetMotorDirection(dir)
			}

			if e.currentFloor != -1 && e.currentFloor == e.targetFloor {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.ClearFloor(e.currentFloor)
				e.ClearButtonLamp(e.currentFloor)
			}
			e.UpdateLastDirection(nextTarget, dir)

		case ES_Moving:
			if e.currentFloor == e.targetFloor {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.ClearFloor(e.currentFloor)
				e.ClearButtonLamp(e.currentFloor)
				e.state = ES_Idle
			} else {
				nextTarget := e.UpdateTargetFloor()
				dir := e.GetMotion()
				elevio.SetMotorDirection(dir)
				e.UpdateLastDirection(nextTarget, dir)
			}		
		case ES_EmergencyStop:
			elevio.SetMotorDirection(elevio.MD_Stop)
			if !ev.EmergencyStop {
                elevio.SetStopLamp(false)
                e.state = ES_Idle
            }
		}
		fmt.Println(e) // DB
	}
}

func (e *Elevator) RunElevatorProgram(port string, id int, numFloors int, initFloor int) {
	// numFloors := 4

	// "localhost:15657"
    elevio.Init("localhost:" + port, numFloors)
    
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

	select{}
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
