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

func (e *Elevator) UpdateMapOfLamps(hallRequests [][2]types.ButtonStatus) {
	for f, row := range hallRequests {
		for b, btnStatus := range row {
			e.UpdateBtnLamp("", btnStatus, f, elevio.ButtonType(b))
		}
	}
}

func (e *Elevator) SetConnectionState(eMsg message.ElevatorMessage) {
	e.mu.Lock()
	e.IsMaster = false
	e.connectedToMaster = true
	e.IsOnline = true
	e.mu.Unlock()
}

func (e *Elevator) TurnOffline() {
	e.mu.Lock()
	e.IsMaster = false
	e.IsOnline = false
	e.connectedToMaster = false
	e.mu.Unlock()
}

func (e *Elevator) TurnToMaster() {
	e.mu.Lock()
	e.IsMaster = true
	e.connectedToMaster = true
	e.mu.Unlock()
}

func (e *Elevator) TurnToSlave() {
	e.mu.Lock()
	e.IsMaster = false
	e.connectedToMaster = true
	e.mu.Unlock()
}
