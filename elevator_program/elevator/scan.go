package elevator

import (
	"elevator_program/elevio"
	"elevator_program/utilities"
)

func (e Elevator) scanFloor(from int, to int, dir elevio.MotorDirection) (bool, elevio.ButtonEvent) {
	numFloors := len(e.system.hallRequests) // Changed to be compatible with system struct

	// saturate bounds
	if from >= numFloors {
		from = numFloors - 1
	}
	if from < 0 {
		from = 0
	}

	if to >= numFloors {
		to = numFloors - 1
	}
	if to < 0 {
		to = 0
	}

	switch dir {
	case elevio.MD_Up:
		for f := from; f <= to; f++ {
			if e.system.Elevators[e.id].CabRequests[f] != NotActive { // Changed to be compatible with system struct
				return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_Cab}

			} else if e.system.hallRequests[f][elevio.BT_HallUp] != NotActive { // Changed to be compatible with system struct
				if e.nextTarget.Floor == f && e.nextTarget.Button == elevio.BT_HallUp { // To not steal anyone elses task
					return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_HallUp}
				}
			}
		}
	case elevio.MD_Down:
		for f := from; f >= to; f-- {
			if e.system.Elevators[e.id].CabRequests[f] != NotActive { // Changed to be compatible with system struct
				return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_Cab}

			} else if e.system.hallRequests[f][elevio.BT_HallDown] != NotActive { // Changed to be compatible with system struct
				if e.nextTarget.Floor == f && e.nextTarget.Button == elevio.BT_HallDown { // To not steal anyone elses task
					return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_HallDown}
				}
			}

		}
	}
	return false, elevio.ButtonEvent{}
}

func (e Elevator) getClosestFloor() elevio.ButtonEvent {
	numFloors := len(e.system.hallRequests) // Changed to be compatible with system struct

	closest := elevio.ButtonEvent{Floor: -1, Button: elevio.BT_Cab}
	minDist := numFloors + 1 // initialize with something bigger than max possible distance
	for f := 0; f < numFloors; f++ {
		dist := utilities.Abs(f - e.currentFloor)

		if e.system.Elevators[e.id].CabRequests[f] == Pending { // Changed to be compatible with system struct, be carefull these might cause error later if emergency stop changes
			if closest.Floor == -1 || dist < minDist {
				closest.Floor = f
				closest.Button = elevio.BT_Cab
				minDist = dist
				continue
			}
		}

		for _, b := range []elevio.ButtonType{elevio.BT_HallUp, elevio.BT_HallDown} {
			if e.system.hallRequests[f][b] == Pending { // Changed to be compatible with system struct, be carefull these might cause error later if emergency stop changes
				if closest.Floor == -1 || dist < minDist {
					closest.Floor = f
					closest.Button = b
					minDist = dist
				}
			}
		}
	}
	return closest
}

func (e Elevator) scanCurrentFloor() (bool, elevio.ButtonEvent) {
	if e.inBetweenFloors {
		return false, elevio.ButtonEvent{}
	}
	return e.scanFloor(e.currentFloor, e.currentFloor, e.direction)
}

func getNextTargetFloor(e Elevator) elevio.ButtonEvent {
	numFloors := len(e.system.hallRequests) // Changed to be compatible with system struct
	bottomFloor := 0
	topFloor := numFloors - 1

	upScan := func() elevio.ButtonEvent {
		if ok, ev := e.scanCurrentFloor(); ok && !e.inBetweenFloors {
			return ev
		}

		// phase 1: continue up
		if ok, ev := e.scanFloor(e.currentFloor+1, topFloor, elevio.MD_Up); ok {
			return ev
		}

		// phase 2: nothing left up, go down
		if ok, ev := e.scanFloor(topFloor, bottomFloor, elevio.MD_Down); ok {
			return ev
		}

		// phase 3: nothing down, move up again
		if ok, ev := e.scanFloor(bottomFloor, e.currentFloor, elevio.MD_Up); ok {
			return ev
		}

		return elevio.ButtonEvent{Floor: -1}
	}

	downScan := func() elevio.ButtonEvent {
		if ok, ev := e.scanCurrentFloor(); ok && !e.inBetweenFloors {
			return ev
		}

		if ok, ev := e.scanFloor(e.currentFloor-1, bottomFloor, elevio.MD_Down); ok {
			return ev
		}
		if ok, ev := e.scanFloor(bottomFloor, topFloor, elevio.MD_Up); ok {
			return ev
		}

		if ok, ev := e.scanFloor(topFloor, e.currentFloor, elevio.MD_Down); ok {
			return ev
		}

		return elevio.ButtonEvent{Floor: -1}
	}
	// endregion

	if e.elevatorState == ES_Idle || e.direction == elevio.MD_Stop {
		return e.getClosestFloor()
	} else if e.direction == elevio.MD_Up {
		return upScan()
	} else if e.direction == elevio.MD_Down {
		return downScan()
	}

	return elevio.ButtonEvent{Floor: -1} // no requests
}
