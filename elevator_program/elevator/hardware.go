package elevator

import (
	"elevator_program/elevio"
)

func (e *Elevator) startButtonPoller() {
	drvButtons := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drvButtons)

	go func() {
		for {
			btn := <-drvButtons
			e.hardwareEventsCh <- HardwareEvent{
				Type:   HW_T_ButtonPress,
				Floor:  btn.Floor,
				Button: btn.Button,
			}
		}
	}()
}

func (e *Elevator) startFloorPoller() {
	drvFloors := make(chan int)
	go elevio.PollFloorSensor(drvFloors)

	go func() {
		for {
			f := <-drvFloors
			e.hardwareEventsCh <- HardwareEvent{
				Type:  HW_T_FloorSensor,
				Floor: f,
			}
		}
	}()
}

func (e *Elevator) startObstructionPoller() {
	drvObstr := make(chan bool)
	go elevio.PollObstructionSwitch(drvObstr)

	go func() {
		for {
			obstr := <-drvObstr
			e.hardwareEventsCh <- HardwareEvent{
				Type:        HW_T_Obstruction,
				Obstruction: obstr,
			}
		}
	}()
}

func (e *Elevator) startStopButtonPoller() {
	drvStop := make(chan bool)
	go elevio.PollStopButton(drvStop)

	go func() {
		for {
			s := <-drvStop
			e.hardwareEventsCh <- HardwareEvent{
				Type:          HW_T_EmergencyStop,
				EmergencyStop: s,
			}
		}
	}()
}

func (e *Elevator) StartHardwareEventsListeners() {
	e.startButtonPoller()
	e.startFloorPoller()
	e.startObstructionPoller()
	e.startStopButtonPoller()
}