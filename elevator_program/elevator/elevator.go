package elevator

import (
	"fmt"

	"elevator_program/utilities"
	"elevator_program/elevio"
)

type Elevator struct {
	id int
	currentFloor int
	targetFloor  int
	floorRequests [][3]bool
	lastMovingDir elevio.MotorDirection
	state    ElevatorState

	// StatusChan chan utilities.StatusMsg
	// TaskChan chan utilities.TaskMsg
}

func (e *Elevator) InitElevator(id int, numFloors int, initTargetFloor int) {
	e.id = id
 	e.targetFloor = initTargetFloor // TODO maybe make dynamic, variable on init
	e.currentFloor = elevio.GetFloor()
	
	e.floorRequests = make([][3]bool, numFloors)

	e.state = ES_Uninitialized
	
	d := e.GetMotion()
    elevio.SetMotorDirection(d)

	// e.state = ES_Moving

	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.targetFloor}
}

func (e Elevator) GetMotion() elevio.MotorDirection {
	if e.targetFloor == -1 || e.currentFloor == e.targetFloor || e.state == ES_EmergencyStop{
		return elevio.MD_Stop
	} else if e.currentFloor < e.targetFloor {
		return elevio.MD_Up
	} else {
		return elevio.MD_Down
	}
}

func (e Elevator) GetNextTargetFloor() elevio.ButtonEvent {
	numFloors := len(e.floorRequests)

	// helper to check if a floor has any of the requested buttons pressed
	hasRequest := func(f int, buttons ...elevio.ButtonType) (bool, elevio.ButtonType) {
		for _, b := range buttons {
			if e.floorRequests[f][b] {
				return true, b
			}
		}
		return false, 0
	}

	upScan := func() elevio.ButtonEvent {
		if ok, btn := hasRequest(e.currentFloor, elevio.BT_Cab, elevio.BT_HallUp); ok { return elevio.ButtonEvent{Floor: e.currentFloor, Button: btn }}

		// phase 1: continue up
		for f := e.currentFloor + 1; f < numFloors; f++ {
			if ok, btn := hasRequest(f, elevio.BT_HallUp, elevio.BT_Cab); ok { return elevio.ButtonEvent{Floor: f, Button: btn } }
		}
		
		// phase 2: nothing left up, go down
		for f := numFloors - 1; f >= 0; f-- {
			if ok, btn := hasRequest(f, elevio.BT_HallDown, elevio.BT_Cab); ok { return elevio.ButtonEvent{Floor: f, Button: btn } }
		}
		
		// phase 3: nothing down, move up again
		for f := 0; f <= e.currentFloor; f++ {
			if ok, btn := hasRequest(f, elevio.BT_HallUp, elevio.BT_Cab); ok { return elevio.ButtonEvent{Floor: f, Button: btn } }
		}

		return elevio.ButtonEvent{Floor: -1}
	}

	downScan := func() elevio.ButtonEvent {
		if ok, btn := hasRequest(e.currentFloor, elevio.BT_Cab, elevio.BT_HallDown); ok {return elevio.ButtonEvent{Floor: e.currentFloor, Button: btn } }

		for f := e.currentFloor - 1; f >= 0; f-- {
			if ok, btn := hasRequest(f, elevio.BT_HallDown, elevio.BT_Cab); ok { return elevio.ButtonEvent{Floor: f, Button: btn } }
		}

		for f := 0; f < numFloors; f++ {
			if ok, btn := hasRequest(f, elevio.BT_HallUp, elevio.BT_Cab); ok { return elevio.ButtonEvent{Floor: f, Button: btn } }
		}

		for f := numFloors - 1; f >= e.currentFloor; f-- {
			if ok, btn := hasRequest(f, elevio.BT_HallDown, elevio.BT_Cab); ok { return elevio.ButtonEvent{Floor: f, Button: btn } }
		}

		return elevio.ButtonEvent{Floor: -1}
	}

	// if elevator is not moving
	if e.state == ES_Idle || e.lastMovingDir == elevio.MD_Stop {
		closest := elevio.ButtonEvent{Floor: -1, Button: elevio.BT_Cab}
		minDist := numFloors + 1 // initialize with something bigger than max possible distance
		for f := 0; f < numFloors; f++ {
			if ok, btn := hasRequest(f, elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab); ok {
				dist := utilities.Abs(f - e.currentFloor)
				if closest.Floor == -1 || dist < minDist {
					closest.Floor = f
					closest.Button = btn
					minDist = dist
				}
			}
		}
		return closest
	}

	if e.lastMovingDir == elevio.MD_Up {
		return upScan()		
	} else if e.lastMovingDir == elevio.MD_Down {
		return downScan()
	}

	return elevio.ButtonEvent{Floor: -1} // no requests
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

func (e *Elevator) ElevatorStateMachine() {
	if e.currentFloor == e.targetFloor {
			fmt.Println("at target floor") // DB
			e.ClearFloor(e.currentFloor)
			e.ClearButtonLamp(e.currentFloor)

			if e.state == ES_Uninitialized {
				e.state = ES_Idle
			}
		}

	if e.state != ES_Uninitialized {
		nextTarget := e.GetNextTargetFloor()

		if nextTarget.Floor == -1 {
			e.state = ES_Idle
			d = elevio.MD_Stop
		} else {
			e.targetFloor = nextTarget.Floor
			e.state = ES_Moving
			
			if nextTarget.Floor == e.currentFloor {
				switch nextTarget.Button {
				case elevio.BT_HallUp:
					e.lastMovingDir = elevio.MD_Up
				case elevio.BT_HallDown:
					e.lastMovingDir = elevio.MD_Down
					}
			}
		}
	}

	d = e.GetMotion()
	fmt.Println("motor is now:", d) // DB
	if d != elevio.MD_Stop {
		e.lastMovingDir = d
	}
	elevio.SetMotorDirection(d)
	
	fmt.Println(e) // DB
}

func (e *Elevator) RunElevatorProgram(port string, id int, numFloors int, initTargetFloor int) {
	// numFloors := 4

	// "localhost:15657"
    elevio.Init("localhost:" + port, numFloors)
    
    var d elevio.MotorDirection = elevio.MD_Up

    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

	e.InitElevator(id, numFloors, initTargetFloor)
    go e.ElevatorStateMachine()
    
    for {
        select {
        case btn := <- drv_buttons:

            // fmt.Printf("%+v\n", btn) // DB
			e.floorRequests[btn.Floor][btn.Button] = true
            elevio.SetButtonLamp(btn.Button, btn.Floor, true)

			if e.state == ES_Idle { e.state = ES_Moving }
            
        case pos := <- drv_floors:
			if pos == -1 { continue }

			e.currentFloor = pos
            
        case a := <- drv_obstr:
            fmt.Printf("%+v\n", a) // DB
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(d)
            }
            
        case a := <- drv_stop:
            fmt.Printf("%+v\n", a) // DB
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                }
            }
        }

		e.ElevatorStateMachine() // TODO not really a state machine ... might become one later
    }    
}

func (e Elevator) String() string {
	s := fmt.Sprintf(
	`Elevator
	id: %d
	current floor: %d
	target floor: %d
	last moving dir: %s
	state: %s
`,
		e.id, e.currentFloor, e.targetFloor, e.lastMovingDir, e.state)

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
