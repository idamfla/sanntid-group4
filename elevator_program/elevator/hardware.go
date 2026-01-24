package elevator

import (
	"elevator_program/elevio"
)

// region startPollers
func (e *Elevator) startButtonPoller() {
	drv_buttons := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drv_buttons)
	go func() {
		for btn := range drv_buttons {
		e.eventsCh <- ElevatorEvent{Type: EV_ButtonPress, Floor: btn.Floor, Button: btn.Button}
		}
	}()
}

func (e *Elevator) startFloorPoller() {
	drv_floors := make(chan int)
	go elevio.PollFloorSensor(drv_floors)
	go func() {
		for f := range drv_floors {
			e.eventsCh <- ElevatorEvent{Type: EV_FloorSensor, Floor: f}
		}
	}()
}

func (e *Elevator) startObstructionPoller() {
	drv_obstr := make(chan bool)
	go elevio.PollObstructionSwitch(drv_obstr)
	go func() {
		for obstr := range drv_obstr {
			e.eventsCh <- ElevatorEvent{Type: EV_Obstruction, Obstruction: obstr}
		}
	}()
}

func (e *Elevator) startStopButtonPoller() {
	drv_stop := make(chan bool)
	go elevio.PollStopButton(drv_stop)
	go func() {
		for s := range drv_stop {
			e.eventsCh <- ElevatorEvent{Type: EV_EmergencyStop, EmergencyStop: s}
		}
	}()
}
// endregion

func (e *Elevator) StartHardwareEventsListeners() {
	e.startButtonPoller()
	e.startFloorPoller()
	e.startObstructionPoller()
	e.startStopButtonPoller()

	done := make(chan struct{})
	<-done
}