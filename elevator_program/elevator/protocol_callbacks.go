package elevator

import (
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
)

func (e *Elevator) ConnectedToMaster() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.connectedToMaster
}

func (e *Elevator) UpdateBtnLamp(ip string, btnStatus types.ButtonStatus, floor int, button elevio.ButtonType) {
	if btnStatus == types.NotActive {
		if button == elevio.BT_Cab {
			if e.Ip == ip {
				e.clearCabLamp(floor)
			}
		} else {
			e.clearHallLamp(floor, button)
		}

	} else {
		if button == elevio.BT_Cab {
			if e.Ip == ip {
				elevio.SetButtonLamp(button, floor, true)
			}
		} else {
			elevio.SetButtonLamp(button, floor, true)
		}
	}
}

func (e *Elevator) SetConnectionState(eMsg message.ElevatorMessage) {
	e.mu.Lock()
	e.Id = eMsg.ID
	e.Ip = eMsg.Addr
	// e.IsMaster = false
	e.connectedToMaster = true
	e.IsOnline = true
	e.mu.Unlock()
}

func (e *Elevator) ClearTarget() {
	clearedTarget := elevio.ButtonEvent{
		Floor:  -1,
		Button: elevio.BT_HallUp,
	}
	e.System.Mutex.Lock()
	elevatorCopy := e.System.Elevators[e.Ip]
	elevatorCopy.Target = clearedTarget
	e.System.Elevators[e.Ip] = elevatorCopy
	e.System.Mutex.Unlock()
}

func (e *Elevator) TurnToMaster() {
	e.mu.Lock()
	// e.IsOnline = true
	e.IsMaster = true
	e.connectedToMaster = true
	e.mu.Unlock()
}

func (e *Elevator) TurnToSlave() {
	e.mu.Lock()
	// e.IsOnline = true
	e.IsMaster = false
	e.connectedToMaster = true
	e.mu.Unlock()
}
