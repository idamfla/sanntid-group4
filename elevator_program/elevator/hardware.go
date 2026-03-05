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
			e.hardwareEventsCh <- HardwareEvent{Type: HW_T_ButtonPress, Floor: btn.Floor, Button: btn.Button}
		}
	}()
}

func (e *Elevator) startFloorPoller() {
	drv_floors := make(chan int)
	go elevio.PollFloorSensor(drv_floors)
	go func() {
		for f := range drv_floors {
			e.hardwareEventsCh <- HardwareEvent{Type: HW_T_FloorSensor, Floor: f}
			e.faultTolerance.FloorEvent()
		}
	}()
}

func (e *Elevator) startObstructionPoller() {
	drv_obstr := make(chan bool)
	go elevio.PollObstructionSwitch(drv_obstr)
	go func() {
		for obstr := range drv_obstr {
			e.hardwareEventsCh <- HardwareEvent{Type: HW_T_Obstruction, Obstruction: obstr}
		}
	}()
}

func (e *Elevator) startStopButtonPoller() {
	drv_stop := make(chan bool)
	go elevio.PollStopButton(drv_stop)
	go func() {
		for s := range drv_stop {
			e.hardwareEventsCh <- HardwareEvent{Type: HW_T_EmergencyStop, EmergencyStop: s}
		}
	}()
}

// endregion

// This function blocks, call in go routine or after everything else is called at start-up
// TODO dont thing the block at the bottom of this function is needed since go routines and channels are not restricted to scopes
func (e *Elevator) StartHardwareEventsListeners() {
	e.startButtonPoller()
	e.startFloorPoller()
	e.startObstructionPoller()
	e.startStopButtonPoller()

	// done := make(chan struct{})
	// <-done
}
