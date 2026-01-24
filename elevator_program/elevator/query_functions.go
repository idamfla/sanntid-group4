package elevator

import (
	"elevator_program/elevio"
	"elevator_program/utilities"
)

func (e Elevator) GetMotion() elevio.MotorDirection {
	if e.targetFloor == -1 || e.currentFloor == e.targetFloor || e.state == ES_EmergencyStop {
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

	upScan := func() elevio.ButtonEvent {
		if ok, btn := hasRequest(e.currentFloor, elevio.BT_Cab, elevio.BT_HallUp); ok {
			return elevio.ButtonEvent{Floor: e.currentFloor, Button: btn}
		}

		// phase 1: continue up
		for f := e.currentFloor + 1; f < numFloors; f++ {
			if ok, btn := hasRequest(f, elevio.BT_HallUp, elevio.BT_Cab); ok {
				return elevio.ButtonEvent{Floor: f, Button: btn}
			}
		}

		// phase 2: nothing left up, go down
		for f := numFloors - 1; f >= 0; f-- {
			if ok, btn := hasRequest(f, elevio.BT_HallDown, elevio.BT_Cab); ok {
				return elevio.ButtonEvent{Floor: f, Button: btn}
			}
		}

		// phase 3: nothing down, move up again
		for f := 0; f <= e.currentFloor; f++ {
			if ok, btn := hasRequest(f, elevio.BT_HallUp, elevio.BT_Cab); ok {
				return elevio.ButtonEvent{Floor: f, Button: btn}
			}
		}

		return elevio.ButtonEvent{Floor: -1}
	}

	downScan := func() elevio.ButtonEvent {
		if ok, btn := hasRequest(e.currentFloor, elevio.BT_Cab, elevio.BT_HallDown); ok {
			return elevio.ButtonEvent{Floor: e.currentFloor, Button: btn}
		}

		for f := e.currentFloor - 1; f >= 0; f-- {
			if ok, btn := hasRequest(f, elevio.BT_HallDown, elevio.BT_Cab); ok {
				return elevio.ButtonEvent{Floor: f, Button: btn}
			}
		}

		for f := 0; f < numFloors; f++ {
			if ok, btn := hasRequest(f, elevio.BT_HallUp, elevio.BT_Cab); ok {
				return elevio.ButtonEvent{Floor: f, Button: btn}
			}
		}

		for f := numFloors - 1; f >= e.currentFloor; f-- {
			if ok, btn := hasRequest(f, elevio.BT_HallDown, elevio.BT_Cab); ok {
				return elevio.ButtonEvent{Floor: f, Button: btn}
			}
		}

		return elevio.ButtonEvent{Floor: -1}
	}

	if e.lastMovingDir == elevio.MD_Up {
		return upScan()
	} else if e.lastMovingDir == elevio.MD_Down {
		return downScan()
	}

	return elevio.ButtonEvent{Floor: -1} // no requests
}
