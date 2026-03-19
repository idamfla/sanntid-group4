package message

import (
	"elevator_program/elevio"
	"elevator_program/types"
)

func NewElevatorMsg(id string, ip string) ElevatorMessage {
	return ElevatorMessage{
		EMsgType: EMSG_T_NewToChannel,
		ID:       id,
		Addr:     ip,
	}
}

func TaskUpdateCallback(id string, task elevio.ButtonEvent, btnStatus types.ButtonStatus) ElevatorMessage {
	return ElevatorMessage{
		EMsgType:  EMSG_T_TaskUpdate,
		ID:        id,
		Task:      task,
		BtnStatus: btnStatus,
	}
}
