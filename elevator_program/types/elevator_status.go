package types

import (
	"elevator_program/elevio"
)

type ElevatorsStatus struct {
	Id             string
	Ip             string
	CurrentFloor   int
	CabRequests    []ButtonStatus
	Target         elevio.ButtonEvent
	Direction      elevio.MotorDirection
	State          ElevatorState
	IsAlive        bool
	IsMotorWorking bool
}
