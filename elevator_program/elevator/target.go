package elevator

import (
	"elevator_program/elevio"
	"elevator_program/types"
	"elevator_program/utilities"
	"fmt"
)

// TODO Sending a pointer should maybe not do that since we don't change the variable

// called by master, e is master, all parameters come from the elevator it checks
func (e *Elevator) isNewTargetBetter(newTarget elevio.ButtonEvent, elev types.ElevatorsStatus) (bool, elevio.ButtonEvent, int) { // TODO can remove task return
	/*
		if dir == md_up && newTarget.Button == bt_down {return false, elevio.ButtonEvent{}}
		else if dir == md_down && newTarget.Button == bt_up

		distTarget int
		distNewTarget int = abs(currFloor-newTarget.Floor)
		if distNewtarget < distTarget {
			return true, newTarget, distNewTarget
		}
	*/

	if elev.State == types.ES_Idle {
		return true, newTarget, utilities.Abs(newTarget.Floor - elev.CurrentFloor)
	}

	switch elev.Direction {
	case elevio.MD_Up:
		if newTarget.Floor < elev.CurrentFloor+1 || newTarget.Button == elevio.BT_HallDown {
			return false, elevio.ButtonEvent{}, len(e.hallRequests) + 1
		}

		// Todo maybe add some logic where we use inBetweenFloor to check if we have gone past currFloor or not??
		distNewTarget := utilities.Abs(newTarget.Floor - (elev.CurrentFloor))
		distTarget := utilities.Abs(elev.Target.Floor - (elev.CurrentFloor))

		if distNewTarget < distTarget {
			return true, newTarget, distNewTarget
		}
	case elevio.MD_Down:
		if newTarget.Floor > elev.CurrentFloor-1 || newTarget.Button == elevio.BT_HallUp {
			return false, elevio.ButtonEvent{}, len(e.hallRequests) + 1
		}

		distNewTarget := utilities.Abs(newTarget.Floor - (elev.CurrentFloor))
		distTarget := utilities.Abs(elev.Target.Floor - (elev.CurrentFloor))

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

func (e *Elevator) ClosestToTarget(elevatorRegistry map[string]types.ElevatorsStatus, newTarget elevio.ButtonEvent) (string, int, elevio.ButtonEvent) {
	minDistance := len(e.hallRequests) + 1
	bestElevatorID := ""
	isClosestIdle := false

	for id, candidate := range elevatorRegistry {
		canTake, _, distance := e.isNewTargetBetter(newTarget, candidate)

		if isClosestIdle && candidate.State != types.ES_Idle {
			continue
		}

		if canTake && distance < minDistance {
			minDistance = distance
			bestElevatorID = id
			isClosestIdle = candidate.State == types.ES_Idle
		}
	}
	return bestElevatorID, minDistance, newTarget
}

// when an elevator asks for a new target
// Todo Do we need currTargetFloor, this one is called when we are looking for a new task??
// func (e *Elevator) ComputeNewTarget(id string, elevator types.ElevatorsStatus, hallRequests [][2]types.ButtonStatus) elevio.ButtonEvent {
// TODO don't need this function
// return e.getNextTargetFloor(elevator, id)
// }

// TODO 'message_handler.go'
/*
 *incomming button press*
	if btn == cab
		if (isNewTargetBetter())
			notifier.nextTarget = newTarget

	else
		elev := ClosesElevator()
*/

func (e *Elevator) IsNewTargetBetterCab(id string, target elevio.ButtonEvent, elevatorStatus types.ElevatorsStatus) bool {
	fmt.Println("Need to debug here: ", e)
	if elevatorStatus.State == types.ES_Idle {
		return true
	}
	fmt.Println("Hihihiha \n\n\n\n\n\n ", target, elevatorStatus)

	if elevatorStatus.Direction == elevio.MD_Up {
		return elevatorStatus.CurrentFloor < target.Floor && target.Floor < elevatorStatus.Target.Floor
	} else {
		return elevatorStatus.Target.Floor < target.Floor && target.Floor < elevatorStatus.CurrentFloor
	}
}
