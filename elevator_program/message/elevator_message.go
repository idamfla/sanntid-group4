package message

import (
	"elevator_program/elevio"
	"elevator_program/types"
)

type ElevatorMessageType int

const (
	EMSG_T_StatusReport ElevatorMessageType = iota
	EMSG_T_StatusReportBroadcast
	EMSG_T_ButtonPress
	EMSG_T_TaskUpdate
	EMSG_T_TaskRequest
	EMSG_T_NewToChannel
	EMSG_T_SyncedElevator
	EMSG_T_IAmMaster
	EMSG_T_SyncSystem
	EMSG_T_IAmAlone
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

func (eMsgType ElevatorMessageType) String() string {
	switch eMsgType {
	case EMSG_T_StatusReport:
		return "Status Report"
	case EMSG_T_StatusReportBroadcast:
		return "Status Report Broadcast"
	case EMSG_T_ButtonPress:
		return "Button Press"
	case EMSG_T_TaskUpdate:
		return "Task Update"
	case EMSG_T_TaskRequest:
		return "Task Request"
	case EMSG_T_NewToChannel:
		return "New to Channel"
	case EMSG_T_SyncedElevator:
		return "Synced elevator"
	case EMSG_T_IAmMaster:
		return "I am master"
	case EMSG_T_SyncSystem:
		return "Sync system"
	case EMSG_T_IAmAlone:
		return "I am alone"
	default:
		return "unknown"
	}
}
