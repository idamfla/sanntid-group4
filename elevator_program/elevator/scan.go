package elevator

import (
	"elevator_program/elevio"
	"elevator_program/utilities"
)

func (e Elevator) scanFloor(from int, to int, so SortingOrder) (bool, elevio.ButtonEvent) {
	numFloors := len(e.floorRequests)

	// saturating range
	if from < 0 {
		from = 0
	}
	if to < 0 {
		to = 0
	}
	if from >= numFloors {
		from = numFloors - 1
	}
	if to >= numFloors {
		to = numFloors - 1
	}

	switch so {
	case SO_Ascending:
		for f := from; f <= to; f++ {
			if e.cabRequests[f] {
				return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_Cab}
			} else if e.floorRequests[f][elevio.BT_HallUp] {
				return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_HallUp}
			}
		}
	case SO_Descending:
		for f := from; f >= to; f-- {
			if e.cabRequests[f] {
				return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_Cab}
			} else if e.floorRequests[f][elevio.BT_HallDown] {
				return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_HallDown}
			}

		}
	}
	return false, elevio.ButtonEvent{}
}

func (e Elevator) scanFromCurrentFloor(so SortingOrder) (bool, elevio.ButtonEvent) {
	start := e.currentFloor
	numFloors := len(e.floorRequests)
	end := numFloors - 1

	switch so {
	case SO_Ascending:
		start = e.currentFloor + 1
		if start > end { // no floors ahead
			return false, elevio.ButtonEvent{}
		}
	case SO_Descending:
		start = e.currentFloor - 1
		if start < 0 { // no floors below
			return false, elevio.ButtonEvent{}
		}
		end = 0
	}

	return e.scanFloor(start, end, so)
}

func (e Elevator) scanCurrentFloor(so SortingOrder) (bool, elevio.ButtonEvent) {
	return e.scanFloor(e.currentFloor, e.currentFloor, so)
}

func (e Elevator) getNextTargetFloor() elevio.ButtonEvent {
	numFloors := len(e.floorRequests)
	bottomFloor := 0
	topFloor := numFloors - 1

	// region scan functions
	defaultScan := func() elevio.ButtonEvent {
		closest := elevio.ButtonEvent{Floor: -1, Button: elevio.BT_Cab}
		minDist := numFloors + 1 // initialize with something bigger than max possible distance
		for f := 0; f < numFloors; f++ {
			dist := utilities.Abs(f - e.currentFloor)

			if e.cabRequests[f] {
				if closest.Floor == -1 || dist < minDist {
					closest.Floor = f
					closest.Button = elevio.BT_Cab
					minDist = dist
					continue
				}
			}

			for _, b := range []elevio.ButtonType{elevio.BT_HallUp, elevio.BT_HallDown} {
				if e.floorRequests[f][b] {
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

	upScan := func() elevio.ButtonEvent {
		if ok, ev := e.scanCurrentFloor(SO_Ascending); ok && !e.inBetweenFloors {
			return ev
		}

		// phase 1: continue up
		if ok, ev := e.scanFromCurrentFloor(SO_Ascending); ok {
			return ev
		}

		// phase 2: nothing left up, go down
		if ok, ev := e.scanFloor(
			topFloor, bottomFloor,
			SO_Descending,
		); ok {
			return ev
		}

		// phase 3: nothing down, move up again
		if ok, ev := e.scanFloor(
			bottomFloor, e.currentFloor,
			SO_Ascending,
		); ok {
			return ev
		}

		return elevio.ButtonEvent{Floor: -1}
	}

	downScan := func() elevio.ButtonEvent {
		if ok, ev := e.scanCurrentFloor(SO_Descending); ok && !e.inBetweenFloors {
			return ev
		}

		if ok, ev := e.scanFromCurrentFloor(SO_Descending); ok {
			return ev
		}

		if ok, ev := e.scanFloor(
			bottomFloor, topFloor,
			SO_Ascending,
		); ok {
			return ev
		}

		if ok, ev := e.scanFloor(
			topFloor, e.currentFloor,
			SO_Descending,
		); ok {
			return ev
		}

		return elevio.ButtonEvent{Floor: -1}
	}
	// endregion

	if e.state == ES_Idle || e.lastDirection == elevio.MD_Stop {
		return defaultScan()
	} else if e.lastDirection == elevio.MD_Up {
		return upScan()
	} else if e.lastDirection == elevio.MD_Down {
		return downScan()
	}

	return elevio.ButtonEvent{Floor: -1} // no requests
}
