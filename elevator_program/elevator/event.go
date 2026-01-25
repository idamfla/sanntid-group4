package elevator

import(
	"elevator_program/elevio"
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

func (e *Elevator) handleEvent(ev ElevatorEvent) {
	switch ev.Type {
		case EV_EmergencyStop:
			elevio.SetStopLamp(ev.EmergencyStop)
			e.emergencyStop = ev.EmergencyStop
		
		case EV_ButtonPress:
			e.floorRequests[ev.Floor][ev.Button] = true
			elevio.SetButtonLamp(ev.Button, ev.Floor, true) // TODO don't turn on lamp before master says to do so
		
		case EV_FloorSensor:
			if ev.Floor != -1 { e.currentFloor = ev.Floor}

		case EV_Obstruction:
			// TODO if door_open, and obsturcion pressed, e.state = obstruction
	}
}