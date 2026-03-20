package elevator

import (
	"elevator_program/elevio"
	"elevator_program/types"
	"elevator_program/utilities"
)

func (e *Elevator) scanFloor(from int, to int, dir elevio.MotorDirection, target elevio.ButtonEvent, hallRequest [][2]types.ButtonStatus, cabRequests []types.ButtonStatus) (bool, elevio.ButtonEvent) {
	numFloors := e.numFloors

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
			if cabRequests[f] != types.NotActive {
				return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_Cab}

			} else if hallRequest[f][elevio.BT_HallUp] != types.NotActive {
				protentilTarget := elevio.ButtonEvent{
					Floor:  f,
					Button: elevio.BT_HallUp,
				}

				if !(target != protentilTarget && hallRequest[f][elevio.BT_HallUp] == types.Running) {
					return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_HallUp}
				}
			}
		}
	case elevio.MD_Down:
		for f := from; f >= to; f-- {
			if cabRequests[f] != types.NotActive {
				return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_Cab}
			} else if hallRequest[f][elevio.BT_HallDown] != types.NotActive {
				protentilTarget := elevio.ButtonEvent{
					Floor:  f,
					Button: elevio.BT_HallDown,
				}
				if !(target != protentilTarget && hallRequest[f][elevio.BT_HallDown] == types.Running) {
					return true, elevio.ButtonEvent{Floor: f, Button: elevio.BT_HallDown}
				}
			}

		}
	}
	return false, elevio.ButtonEvent{}
}

func (e *Elevator) getClosestFloor(elevator types.ElevatorsStatus, hallRequests [][2]types.ButtonStatus) elevio.ButtonEvent {
	numFloors := len(hallRequests)

	e.mu.Lock()
	floor := e.currentFloor
	e.mu.Unlock()

	closest := elevio.ButtonEvent{Floor: -1, Button: elevio.BT_Cab}
	minDist := numFloors + 1
	for f := 0; f < numFloors; f++ {
		dist := utilities.Abs(f - floor)

		if elevator.CabRequests[f] != types.NotActive {
			if closest.Floor == -1 || dist < minDist {
				closest.Floor = f
				closest.Button = elevio.BT_Cab
				minDist = dist
				continue
			}
		}

		for _, b := range []elevio.ButtonType{elevio.BT_HallUp, elevio.BT_HallDown} {
			if hallRequests[f][b] != types.NotActive {
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

func (e *Elevator) scanCurrentFloor() (bool, elevio.ButtonEvent) {
	e.mu.Lock()
	between := e.inBetweenFloors
	e.mu.Unlock()
	if between {
		return false, elevio.ButtonEvent{}
	}
	e.System.Mutex.RLock()
	direction := e.System.Elevators[e.Id].Direction
	cabRequests := e.System.Elevators[e.Id].CabRequests
	hallRequest := e.System.HallRequests
	target := e.System.Elevators[e.Id].Target
	e.System.Mutex.RUnlock()

	e.mu.Lock()
	floor := e.currentFloor
	e.mu.Unlock()
	return e.scanFloor(floor, floor, direction, target, hallRequest, cabRequests)
}

func (e *Elevator) GetNextTargetFloor(elevator types.ElevatorsStatus, hallRequests [][2]types.ButtonStatus) elevio.ButtonEvent {

	numFloors := len(hallRequests)
	bottomFloor := 0
	topFloor := numFloors - 1

	target := elevator.Target

	upScan := func() elevio.ButtonEvent {
		e.mu.Lock()
		between := e.inBetweenFloors
		e.mu.Unlock()
		if ok, ev := e.scanCurrentFloor(); ok && !between {
			return ev
		}

		if ok, ev := e.scanFloor(elevator.CurrentFloor+1, topFloor, elevio.MD_Up, target, hallRequests, elevator.CabRequests); ok {
			return ev
		}

		if ok, ev := e.scanFloor(topFloor, bottomFloor, elevio.MD_Down, target, hallRequests, elevator.CabRequests); ok {
			return ev
		}

		if ok, ev := e.scanFloor(bottomFloor, elevator.CurrentFloor, elevio.MD_Up, target, hallRequests, elevator.CabRequests); ok {
			return ev
		}

		return elevio.ButtonEvent{Floor: -1}
	}

	downScan := func() elevio.ButtonEvent {
		e.mu.Lock()
		between := e.inBetweenFloors
		e.mu.Unlock()
		if ok, ev := e.scanCurrentFloor(); ok && !between {
			return ev
		}

		if ok, ev := e.scanFloor(elevator.CurrentFloor-1, bottomFloor, elevio.MD_Down, target, hallRequests, elevator.CabRequests); ok {
			return ev
		}
		if ok, ev := e.scanFloor(bottomFloor, topFloor, elevio.MD_Up, target, hallRequests, elevator.CabRequests); ok {
			return ev
		}

		if ok, ev := e.scanFloor(topFloor, elevator.CurrentFloor, elevio.MD_Down, target, hallRequests, elevator.CabRequests); ok {
			return ev
		}

		return elevio.ButtonEvent{Floor: -1}
	}

	if elevator.State == types.ES_Idle || elevator.Direction == elevio.MD_Stop {
		return e.getClosestFloor(elevator, hallRequests)
	} else if elevator.Direction == elevio.MD_Up {
		return upScan()
	} else if elevator.Direction == elevio.MD_Down {
		return downScan()
	}
	return elevio.ButtonEvent{Floor: -1}
}
