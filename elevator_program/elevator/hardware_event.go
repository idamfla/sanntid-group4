package elevator

import (
	"elevator_program/elevio"
	"fmt"
)

type HardwareType int

const (
	HW_T_FloorSensor HardwareType = iota
	HW_T_ButtonPress
	HW_T_Obstruction
	HW_T_EmergencyStop
)

type HardwareEvent struct {
	Type          HardwareType
	Floor         int
	Button        elevio.ButtonType
	Obstruction   bool
	EmergencyStop bool
}

func (e *Elevator) handleHardwareEvent(hwEvent HardwareEvent) {
	switch hwEvent.Type {
	case HW_T_EmergencyStop:
		elevio.SetStopLamp(hwEvent.EmergencyStop)
		e.emergencyStop = hwEvent.EmergencyStop

	case HW_T_ButtonPress:
        if e.offline && hwEvent.Button != elevio.BT_Cab {
                elevio.SetButtonLamp(hwEvent.Button, hwEvent.Floor, false)
                break
            }

		if hwEvent.Button == elevio.BT_Cab {
			e.cabRequests[hwEvent.Floor] = true
		} else {
			e.hallRequests[hwEvent.Floor][hwEvent.Button] = Pending
		}
		elevio.SetButtonLamp(hwEvent.Button, hwEvent.Floor, true) // TODO don't turn on lamp before master says to do so

	case HW_T_FloorSensor:
		if hwEvent.Floor == -1 {
			e.inBetweenFloors = true // TODO maybe set inBetweenFloors true when the elevator moves, not when we arrive at correct floor
		} else {
			elevio.SetFloorIndicator(hwEvent.Floor)
			e.currentFloor = hwEvent.Floor
			e.inBetweenFloors = false
		}

	case HW_T_Obstruction:
		if e.doorState == DS_Closed {
			return
		}
		e.obstruction = hwEvent.Obstruction
	}
}

func (e *Elevator) RunHardwareEventLoop() {
	fmt.Println("EVENT LOOP STARTED")
	for hwEvent := range e.hardwareEventsCh {
		e.handleHardwareEvent(hwEvent)
		fmt.Println(e) // DB
	}
}
