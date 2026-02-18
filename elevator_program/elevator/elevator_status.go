package elevator

import "elevator_program/elevio"

type ElevatorsStatus struct {
	id           int
	currentFloor int
	cabRequests  []bool
	target       elevio.ButtonEvent
	direction    elevio.MotorDirection
	state        ElevatorState
}
