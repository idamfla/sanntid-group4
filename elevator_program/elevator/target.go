package elevator

import "elevator_program/elevio"

// called by master, e is master, all parameters come from the elevator it checks
func (e Elevator) isNewTargetBetter(currFloor int, currTarget elevio.ButtonEvent, newTarget elevio.ButtonEvent, dir elevio.MotorDirection) (bool, elevio.ButtonEvent, int) {
	/*
		if dir == md_up && newTarget.Button == bt_down {return false, elevio.ButtonEvent{}}
		else if dir == md_down && newTarget.Button == bt_up

		distTarget int
		distNewTarget int = abs(currFloor-newTarget.Floor)
		if distNewtarget < distTarget {
			return true, newTarget, distNewTarget
		}
	*/
	return false, elevio.ButtonEvent{}, len(e.hallRequests) + 1
}

/* have funciton that runs isNewTargetBetter on all elevators in the elevator
map (it is a map, not 2d arr), use the output dist to find who is closest and
let them do the task. this to avoid the first elevator in the map to always take
on the extra work if the new task is better for it but still another one is closer
*/

// when an elevator asks for a new target
func (e Elevator) computeNewTarget(currFloor int, currTargetFloor int, cabRequests []bool, dir elevio.MotorDirection) elevio.ButtonEvent {
	// ruffly same as the basic scan_logic
	return elevio.ButtonEvent{}
}
