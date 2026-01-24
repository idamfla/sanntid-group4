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
	Type          EventType
	Floor         int
	Button        elevio.ButtonType
	Obstruction   bool
	EmergencyStop bool
}

func (s ElevatorState) String() string {
	switch s {
	// case Idle:
	// 		return "idle"
	case ES_Uninitialized:
		return "uninitialized"
	case ES_Idle:
		return "idle"
	case ES_Moving:
		return "moving"
	case ES_DoorOpen:
		return "door open"
	case ES_Obstruction:
		return "obstruction"
	case ES_EmergencyStop:
		return "emergency stop"
	default:
		return "unknown"
	}
}
