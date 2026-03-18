package elevator

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
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
	if e.IsOnline {
		// TODO Temp for debugging
		e.connectedToMaster = true
		e.handleHardwareEventOnline(hwEvent)
	} else {
		e.handleHardwareEventOffline(hwEvent)
	}
}

func (e *Elevator) handleHardwareEventOnline(hwEvent HardwareEvent) {
	switch hwEvent.Type {
	case HW_T_EmergencyStop:
		elevio.SetStopLamp(hwEvent.EmergencyStop)
		e.emergencyStop = hwEvent.EmergencyStop
		e.System.Mutex.RLock()
		_, elevators := e.System.Snapshot()
		e.System.Mutex.RUnlock()
		msg := message.ElevatorMessage{
			MsgType:   types.MSG_T_StatusReport,
			Id:        e.Id,
			Elevators: elevators,
		}
		e.SendToCoordinator <- msg

	case HW_T_ButtonPress:
		if !e.connectedToMaster {
			println("Not connected to master, cannot accept buttonpress")
			return
		}
		task := elevio.ButtonEvent{
			Floor:  hwEvent.Floor,
			Button: hwEvent.Button,
		}

		// Check if button press is already in system, no need to message master
		if !e.System.IsRequestInSystem(e.Id, task) {
			e.System.Mutex.RLock()
			_, elevators := e.System.Snapshot()
			e.System.Mutex.RUnlock()
			msg := message.ElevatorMessage{
				MsgType:   types.MSG_T_ButtonPress,
				Id:        e.Id,
				Task:      task,
				BtnStatus: types.Pending,
				Elevators: map[string]types.ElevatorsStatus{
					e.Id: elevators[e.Id],
				},
			}

			if e.IsMaster {
				taskElevatorId, _, _ := e.ClosestToTarget(elevators, task)
				if taskElevatorId != e.Id {
					msg.BtnStatus = types.Running
					msg.Id = taskElevatorId
				}
			}
			e.SendToCoordinator <- msg
		}

	case HW_T_FloorSensor:
		if hwEvent.Floor == -1 {
			e.inBetweenFloors = true // TODO maybe set inBetweenFloors true when the elevator moves, not when we arrive at correct floor
		} else {
			elevio.SetFloorIndicator(hwEvent.Floor)
			e.currentFloor = hwEvent.Floor
			e.inBetweenFloors = false

			e.System.Mutex.Lock()
			elevatorCopy := e.System.Elevators[e.Id]
			elevatorCopy.CurrentFloor = hwEvent.Floor
			e.System.Elevators[e.Id] = elevatorCopy
			msg := message.ElevatorMessage{
				MsgType:   types.MSG_T_StatusReport,
				Id:        e.Id,
				Elevators: e.System.Elevators,
			}
			e.SendToCoordinator <- msg
			e.System.Mutex.Unlock()
		}

	case HW_T_Obstruction:
		if e.doorState == DS_Closed {
			return
		}
		e.obstruction = hwEvent.Obstruction // TODO should i notify master, probably not right?
	}
}

func (e *Elevator) handleHardwareEventOffline(hwEvent HardwareEvent) {
	switch hwEvent.Type {
	case HW_T_EmergencyStop:
		elevio.SetStopLamp(hwEvent.EmergencyStop)
		e.emergencyStop = hwEvent.EmergencyStop

	case HW_T_ButtonPress:
		if hwEvent.Button == elevio.BT_Cab {
			e.System.Mutex.Lock()
			e.System.Elevators[e.Id].CabRequests[hwEvent.Floor] = types.Pending // Changed to be compatible with system struct
			e.System.Mutex.Unlock()
		} else {
			fmt.Println("Elevator is offline, can not accept order")
			return
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
