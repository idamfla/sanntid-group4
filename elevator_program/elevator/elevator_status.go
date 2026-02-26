package elevator

import "elevator_program/elevio"

type ElevatorsStatus struct {
	id            int
	currentFloor  int
	cabRequests   []bool
	target        elevio.ButtonEvent
	direction     elevio.MotorDirection
	elevatorState ElevatorState
	// temp, just need to know the ip
	ip string
}
