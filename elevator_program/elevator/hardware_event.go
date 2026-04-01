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
	e.mu.Lock()
	online := e.IsOnline
	if online {
		e.connectedToMaster = true
	}
	e.mu.Unlock()

	if online {
		e.handleHardwareEventOnline(hwEvent)
	} else {
		e.handleHardwareEventOffline(hwEvent)
	}
}

func (e *Elevator) handleHardwareEventOnline(hwEvent HardwareEvent) {
	switch hwEvent.Type {
	case HW_T_EmergencyStop:
		elevio.SetStopLamp(hwEvent.EmergencyStop)
		e.mu.Lock()
		e.emergencyStop = hwEvent.EmergencyStop
		e.mu.Unlock()
		e.System.Mutex.Lock()
		elevatorCopy := e.System.Elevators[e.Ip]
		elevatorCopy.State = types.ES_EmergencyStop
		e.System.Elevators[e.Ip] = elevatorCopy
		eMsg := message.ElevatorMessage{
			EMsgType:  message.EMSG_T_StatusReport,
			ID:        e.Id,
			Addr:      e.Ip,
			Elevators: e.System.Elevators,
		}
		e.System.Mutex.Unlock()
		e.SendToCoordinator <- eMsg

	case HW_T_ButtonPress:
		e.mu.Lock()
		connected := e.connectedToMaster
		e.mu.Unlock()
		if !connected {
			println("Not connected to master, cannot accept buttonpress")
			return
		}
		task := elevio.ButtonEvent{
			Floor:  hwEvent.Floor,
			Button: hwEvent.Button,
		}

		if !e.System.IsRequestInSystem(e.Ip, task) {
			e.System.Mutex.RLock()
			_, elevators := e.System.Snapshot()
			e.System.Mutex.RUnlock()
			eMsg := message.ElevatorMessage{
				EMsgType:  message.EMSG_T_ButtonPress,
				ID:        e.Id,
				Addr:      e.Ip,
				Task:      task,
				BtnStatus: types.Pending,
				Elevators: map[string]types.ElevatorsStatus{
					e.Ip: elevators[e.Ip],
				},
			}

			if e.IsMaster {
				taskElevatorIp := e.ClosestToTarget(elevators, task)
				if taskElevatorIp != e.Ip {
					eMsg.BtnStatus = types.Running
					eMsg.Addr = taskElevatorIp
				}
			}
			e.SendToCoordinator <- eMsg
		}

	case HW_T_FloorSensor:
		if hwEvent.Floor == -1 {
			e.mu.Lock()
			e.inBetweenFloors = true
			e.mu.Unlock()
		} else {
			elevio.SetFloorIndicator(hwEvent.Floor)
			e.mu.Lock()
			e.currentFloor = hwEvent.Floor
			e.inBetweenFloors = false
			e.mu.Unlock()

			e.System.Mutex.Lock()
			elevatorCopy := e.System.Elevators[e.Ip]
			elevatorCopy.CurrentFloor = hwEvent.Floor
			e.System.Elevators[e.Ip] = elevatorCopy
			_, elevs := e.System.Snapshot()
			e.System.Mutex.Unlock()

			eMsg := message.ElevatorMessage{
				EMsgType:  message.EMSG_T_StatusReport,
				ID:        e.Id,
				Addr:      e.Ip,
				Elevators: elevs,
			}
			e.SendToCoordinator <- eMsg
		}

	case HW_T_Obstruction:
		e.mu.Lock()
		closed := e.doorState == DS_Closed
		e.mu.Unlock()
		if closed {
			return
		}
		e.mu.Lock()
		e.obstruction = hwEvent.Obstruction
		e.mu.Unlock()
	}
}

func (e *Elevator) handleHardwareEventOffline(hwEvent HardwareEvent) {
	switch hwEvent.Type {
	case HW_T_EmergencyStop:
		elevio.SetStopLamp(hwEvent.EmergencyStop)
		e.mu.Lock()
		e.emergencyStop = hwEvent.EmergencyStop
		e.mu.Unlock()

	case HW_T_ButtonPress:
		if hwEvent.Button == elevio.BT_Cab {
			e.System.Mutex.Lock()
			e.System.Elevators[e.Ip].CabRequests[hwEvent.Floor] = types.Pending
			e.System.Mutex.Unlock()
		} else {
			fmt.Println("Elevator is offline, can not accept order")
			return
		}
		elevio.SetButtonLamp(hwEvent.Button, hwEvent.Floor, true)

	case HW_T_FloorSensor:
		if hwEvent.Floor == -1 {
			e.mu.Lock()
			e.inBetweenFloors = true
			e.mu.Unlock()
		} else {
			elevio.SetFloorIndicator(hwEvent.Floor)
			e.mu.Lock()
			e.currentFloor = hwEvent.Floor
			e.inBetweenFloors = false
			e.mu.Unlock()

			e.System.Mutex.Lock()
			elevatorCopy := e.System.Elevators[e.Ip]
			elevatorCopy.CurrentFloor = hwEvent.Floor
			e.System.Elevators[e.Ip] = elevatorCopy
			e.System.Mutex.Unlock()
		}

	case HW_T_Obstruction:
		e.mu.Lock()
		closed := e.doorState == DS_Closed
		e.mu.Unlock()
		if closed {
			return
		}
		e.mu.Lock()
		e.obstruction = hwEvent.Obstruction
		e.mu.Unlock()
	}
}

func (e *Elevator) RunHardwareEventLoop() {
	defer e.wg.Done()
	fmt.Println("EVENT LOOP STARTED")
	for {
		select {
		case <-e.stop:
			return
		case hwEvent := <-e.hardwareEventsCh:
			e.handleHardwareEvent(hwEvent)
		}
	}
}
