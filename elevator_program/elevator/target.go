package elevator

import (
	"elevator_program/elevio"
	"elevator_program/utilities"
)

// called by master, e is master, all parameters come from the elevator it checks
func (e Elevator) isNewTargetBetter(newTarget elevio.ButtonEvent, elev ElevatorsStatus) (bool, elevio.ButtonEvent, int) {
	/*
		if dir == md_up && newTarget.Button == bt_down {return false, elevio.ButtonEvent{}}
		else if dir == md_down && newTarget.Button == bt_up

		distTarget int
		distNewTarget int = abs(currFloor-newTarget.Floor)
		if distNewtarget < distTarget {
			return true, newTarget, distNewTarget
		}
	*/

	if elev.state == ES_Idle {
		return true, newTarget, utilities.Abs(newTarget.Floor - elev.currentFloor)
	}

	switch elev.direction {
	case elevio.MD_Up:
		if newTarget.Floor < elev.currentFloor+1 || newTarget.Button == elevio.BT_HallDown {
			return false, elevio.ButtonEvent{}, len(e.hallRequests) + 1
		}

		// Todo maybe add some logic where we use inBetweenFloor to check if we have gone past currFloor or not??
		distNewTarget := utilities.Abs(newTarget.Floor - (elev.currentFloor + 1))
		distTarget := utilities.Abs(elev.target.Floor - (elev.currentFloor + 1))

		if distNewTarget < distTarget {
			return true, newTarget, distNewTarget
		}
	case elevio.MD_Down:
		if newTarget.Floor > elev.currentFloor-1 || newTarget.Button == elevio.BT_HallUp {
			return false, elevio.ButtonEvent{}, len(e.hallRequests) + 1
		}

		distNewTarget := utilities.Abs(newTarget.Floor - (elev.currentFloor - 1))
		distTarget := utilities.Abs(elev.target.Floor - (elev.currentFloor - 1))

		if distNewTarget < distTarget {
			return true, newTarget, distNewTarget
		}
	}

	return false, elevio.ButtonEvent{}, len(e.hallRequests) + 1 // Should I send just elevio.ButtonEvent since it does not matter
}

/* have funciton that runs isNewTargetBetter on all elevators in the elevator
map (it is a map, not 2d arr), use the output dist to find who is closest and
let them do the task. this to avoid the first elevator in the map to always take
on the extra work if the new task is better for it but still another one is closer
*/

func (e Elevator) closestToTarget(elevatorRegistry map[int]ElevatorsStatus, newTarget elevio.ButtonEvent) (int, int, elevio.ButtonEvent) {
	dist := len(e.hallRequests)
	closest := -1
	isClosestIdle := false

	for id, elev := range elevatorRegistry {
		ok, _, distance := e.isNewTargetBetter(newTarget, elev)

		if isClosestIdle && elev.state != ES_Idle {
			continue
		}

		if ok && distance < dist {
			dist = distance
			closest = id
			ok = false
			isClosestIdle = elev.state == ES_Idle
		}
	}
	return closest, dist, newTarget
}

// when an elevator asks for a new target
func (e Elevator) computeNewTarget(currFloor int, currTargetFloor int, cabRequests []bool, dir elevio.MotorDirection) elevio.ButtonEvent {
	// ruffly same as the basic scan_logic
	return elevio.ButtonEvent{}
}
