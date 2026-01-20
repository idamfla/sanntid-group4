package elevator

import (
	"fmt"

	// "elevator_program/utilities"
	"elevator_program/elevio"
)

var _numFloors int = 4 // TODO integrate

type Elevator struct {
	id int
	currentFloor int
	targetFloor  int
	currentPosition int
	floorRequests [][3]bool
	state    ElevatorState

	// StatusChan chan utilities.StatusMsg
	// TaskChan chan utilities.TaskMsg
}

func (e *Elevator) InitElevator(id int, numFloors int) {
	e.id = id
	e.currentPosition = elevio.GetFloor()
	e.targetFloor = 0 // TODO maybe make dynamic, variable on init
	e.currentFloor = elevio.GetFloor()
	
	e.floorRequests = make([][]bool, numFloors)
    for f := 0; f < numFloors; f++ {
        e.floorRequests[f] = make([]bool, 3) // 3 ButtonTypes
    }

	e.state = ES_Uninitialized
	
	d := e.GetMotion()
    elevio.SetMotorDirection(d)

	// e.state = ES_Running

	// e.StatusChan = statusChan
	// e.TaskChan = taskChan

	// e.StatusChan <-utilities.StatusMsg{e.id, e.currentFloor, e.targetFloor}
}

func (e Elevator) GetMotion() elevio.MotorDirection {
	if e.targetFloor == -1 || e.currentPosition == e.targetFloor {
		return elevio.MD_Stop
	} else if e.currentFloor < e.targetFloor {
		return elevio.MD_Up
	} else {
		return elevio.MD_Down
	}
}

func (e Elevator) GetNextTargetFloor() int {
	// var d elevio.MotorDirection = e.GetMotion()
	// numFloors := len(e.floorRequests)
	
	// if d == elevio.MD_Up {
	// 	for f := e.currentFloor + 1; f < numFloors; f++ {
	// 		if e.floorRequests[f][BT_HallUp] || e.floorRequests[f][BT_Cab] {
	// 			return f
	// 		}
	// 	}

	// 	for f := numFloors - 1; f >= 0; f-- {
	// 		if e.floorRequests[f][BT_HallDown] || e.floorRequests[f][BT_Cab] {
	// 			return f
	// 		}
	// 	}

	// 	for f := 0; f < e.currentFloor; f++ {
	// 		if e.floorRequests[f][BT_HallUp] || e.floorRequests[f][BT_Cab] {
	// 			return f
	// 		}
	// 	}
	// } else if d == elevio.MD_Down {
	// 	for f := e.currentFloor - 1; f >= 0; f-- {
	// 		if e.floorRequests[f][BT_HallDown] || e.floorRequests[f][BT_Cab] {
	// 			return f
	// 		}
	// 	}
		
	// 	for f := 0; f < numFloors; f++ {
	// 		if e.floorRequests[f][BT_HallUp] || e.floorRequests[f][BT_Cab] {
	// 			return f
	// 		}
	// 	}

	// 	for f := numFloors - 1; f >= 0; f-- {
	// 		if e.floorRequests[f][BT_HallDown] || e.floorRequests[f][BT_Cab] {
	// 			return f
	// 		}
	// 	}
	// }

	d := e.GetMotion()
	numFloors := len(e.floorRequests)

	// Helper to check if a floor has any of the requested buttons pressed
	hasRequest := func(f int, buttons ...elevio.ButtonType) bool {
		for _, b := range buttons {
			if e.floorRequests[f][b] {
				return true
			}
		}
		return false
	}

	// Priority 1: Floors ahead in same direction (Up/Down) or Cab stops
	if d == elevio.MD_Up {
		for f := e.currentFloor + 1; f < numFloors; f++ {
			if hasRequest(f, BT_HallUp, BT_Cab) {
				return f
			}
		}
	} else if d == elevio.MD_Down {
		for f := e.currentFloor - 1; f >= 0; f-- {
			if hasRequest(f, BT_HallDown, BT_Cab) {
				return f
			}
		}
	}

	// Priority 2: Floors behind in opposite direction or Cab stops
	for f := 0; f < numFloors; f++ {
		if d == elevio.MD_Up && f <= e.currentFloor {
			if hasRequest(f, BT_HallDown, BT_Cab) {
				return f
			}
		} else if d == elevio.MD_Down && f >= e.currentFloor {
			if hasRequest(f, BT_HallUp, BT_Cab) {
				return f
			}
		}
	}

	// Priority 3: Remaining requests in the same direction (less urgent)
	if d == elevio.MD_Up {
		for f := e.currentFloor + 1; f < numFloors; f++ {
			if hasRequest(f, BT_HallDown) {
				return f
			}
		}
	} else if d == elevio.MD_Down {
		for f := e.currentFloor - 1; f >= 0; f-- {
			if hasRequest(f, BT_HallUp) {
				return f
			}
		}
	}

	return e.currentFloor // no requests
}

func (e *Elevator) RunElevatorProgram(port string, id int, numFloors int) {
	// numFloors := 4

	// "localhost:15657"
    elevio.Init("localhost:" + port, numFloors)
    
    var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)
    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

	e.InitElevator(id)
    
    
    for {
        select {
        case btn := <- drv_buttons:
			if e.state == ES_Uninitialized { break }

            fmt.Printf("%+v\n", btn)
			e.floorRequests[btn.Floor][btn.Button] = true
            elevio.SetButtonLamp(btn.Button, btn.Floor, true)
            
        case pos := <- drv_floors:
			e.currentPosition = pos
			if e.currentPosition != -1 {
				e.currentFloor = pos
				
				if e.currentFloor == e.targetFloor && e.state == ES_Uninitialized {
					e.state = ES_Running
				}
			}

			e.GetNextTargetFloor()
			d = e.GetMotion()
            // fmt.Printf("%+v\n", a)
            // if a == numFloors-1 {
            //     d = elevio.MD_Down
            // } else if a == 0 {
            //     d = elevio.MD_Up
            // }
            elevio.SetMotorDirection(d)
			fmt.Println(e)
            
            
        case a := <- drv_obstr:
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(d)
            }
            
        case a := <- drv_stop:
            fmt.Printf("%+v\n", a)
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                }
            }
        }
    }    
}

func (e Elevator) String() string {
	return fmt.Sprintf(`Elevator
	id: %d
	current position: %d
	current floor: %d
	target floor: %d
	state: %s`,
		e.id, e.currentFloor, e.currentFloor, e.targetFloor, e.state)
}
