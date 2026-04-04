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
	EMSG_T_SyncedElevator // TODO Is supposed to don't do shit
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
	case EMSG_T_TaskCreate:
		return "Task Create"
	case EMSG_T_ButtonPress:
		return "Button Press"
	// EMSG_T_TaskAssign   // a task is assigned to you
	// EMSG_T_TaskDelegate // a task is assigned to another person
	case EMSG_T_TaskUpdate:
		return "Task Update"
	case EMSG_T_TaskComplete:
		return "Task Complete"
	case EMSG_T_TaskRequest:
		return "Task Request"
	case EMSG_T_LostComs:
		return "Lost Conn" // "Coms" or "Conn"?
	case EMSG_T_ElevatorLost:
		return "Elevator Lost"
	case EMSG_T_NewToChannel:
		return "New to Channel"
	default:
		return "unknown"
	}
}
