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
	minDistance := len(e.hallRequests) + 1
	bestElevatorID := -1
	isClosestIdle := false

	for id, candidate := range elevatorRegistry {
		canTake, _, distance := e.isNewTargetBetter(newTarget, candidate)

		if isClosestIdle && candidate.state != ES_Idle {
			continue
		}

		if canTake && distance < minDistance {
			minDistance = distance
			bestElevatorID = id
			isClosestIdle = candidate.state == ES_Idle
		}
	}
	return bestElevatorID, minDistance, newTarget
}

// when an elevator asks for a new target
// Todo Do we need currTargetFloor, this one is called when we are looking for a new task??
func (e Elevator) computeNewTarget(currFloor int, cabRequests []bool, dir elevio.MotorDirection) elevio.ButtonEvent {
	// ruffly same as the basic scan_logic

	// Hallrequests need to change its logic, pending, active, ...
	hallRequestsCopy := e.hallRequests // Needs to be sure we don't modify e.hallRequests
	elevatorCopy := Elevator{
		hallRequests: hallRequestsCopy,
		currentFloor: currFloor,
		cabRequests:  cabRequests,
		direction:    dir,
	}

	return getNextTargetFloor(elevatorCopy)
}
