package elevator

import (
	"elevator_program/elevio"
	"elevator_program/types"
	"elevator_program/utilities"
)

func (e *Elevator) isNewTargetBetter(newTarget elevio.ButtonEvent, elev types.ElevatorsStatus) (bool, int) {
	if elev.State == types.ES_Idle {
		return true, utilities.Abs(newTarget.Floor - elev.CurrentFloor)
	}

	switch elev.Direction {
	case elevio.MD_Up:
		if newTarget.Floor < elev.CurrentFloor+1 || newTarget.Button == elevio.BT_HallDown {
			return false, e.NumFloors + 1
		}

		distNewTarget := newTarget.Floor - (elev.CurrentFloor)
		distTarget := elev.Target.Floor - (elev.CurrentFloor)

		if distNewTarget < distTarget && distNewTarget > 0 {
			return true, distNewTarget
		}
	case elevio.MD_Down:
		if newTarget.Floor > elev.CurrentFloor-1 || newTarget.Button == elevio.BT_HallUp {
			return false, e.NumFloors + 1
		}

		distNewTarget := (elev.CurrentFloor) - newTarget.Floor
		distTarget := (elev.CurrentFloor) - elev.Target.Floor

		if distNewTarget < distTarget && distNewTarget > 0 {
			return true, distNewTarget
		}
	}

	return false, e.NumFloors + 1
}

func (e *Elevator) ClosestToTarget(elevatorRegistry map[string]types.ElevatorsStatus, newTarget elevio.ButtonEvent) string {
	minDistance := len(e.System.HallRequests) + 1
	bestElevatorIP := ""
	isClosestIdle := false

	for ip, candidate := range elevatorRegistry {
		canTake, distance := e.isNewTargetBetter(newTarget, candidate)

		if isClosestIdle && candidate.State != types.ES_Idle {
			continue
		}

		if canTake && distance < minDistance {
			minDistance = distance
			bestElevatorIP = ip
			isClosestIdle = candidate.State == types.ES_Idle
		}
	}
	return bestElevatorIP
}

func (e *Elevator) IsNewTargetBetterCab(target elevio.ButtonEvent, elevatorStatus types.ElevatorsStatus) bool {
	if elevatorStatus.State == types.ES_Idle {
		return true
	}

	if elevatorStatus.Direction == elevio.MD_Up {
		return elevatorStatus.CurrentFloor < target.Floor && target.Floor < elevatorStatus.Target.Floor
	} else {
		return elevatorStatus.Target.Floor < target.Floor && target.Floor < elevatorStatus.CurrentFloor
	}
}
