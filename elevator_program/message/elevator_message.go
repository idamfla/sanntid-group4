package message

import (
	"elevator_program/elevio"
	"elevator_program/types"
)

type ElevatorMessageType int

const (
	EMSG_T_StatusReport ElevatorMessageType = iota
	EMSG_T_StatusReportBroadcast
	EMSG_T_TaskCreate
	EMSG_T_ButtonPress
	EMSG_T_TaskUpdate
	EMSG_T_TaskComplete
	EMSG_T_TaskRequest
	EMSG_T_LostComs
	EMSG_T_ElevatorLost
	EMSG_T_NewToChannel
)

type ElevatorMessage struct {
	EMsgType ElevatorMessageType

	ID   string
	Addr string

	ActivePeers int
	Task        elevio.ButtonEvent
	BtnStatus   types.ButtonStatus

	HallRequests [][2]types.ButtonStatus
	Elevators    map[string]types.ElevatorsStatus
}
