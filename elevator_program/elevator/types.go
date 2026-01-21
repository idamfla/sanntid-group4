package elevator

import (
	"elevator_program/elevio"
)

type ElevatorState int

const (
	ES_Uninitialized ElevatorState = iota
	ES_Idle
	ES_Moving
	ES_DoorOpen
	ES_Obstruction
	ES_EmergencyStop
)

type EventType int

const (
	EV_FloorSensor EventType = iota
	EV_ButtonPress
	EV_Obstruction
	EV_EmergencyStop
	EV_TaskAssigned
	EV_TaskCompleted
)

type ElevatorEvent struct {
	Type EventType
	Floor int
	Button elevio.ButtonType
	Obstruction bool
	EmergencyStop bool
}