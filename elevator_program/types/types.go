package types

type ElevatorState int

const (
	ES_Uninitialized ElevatorState = iota
	ES_Idle
	ES_Moving
	ES_DoorOpen
	// ES_Obstruction // TODO move to DoorState
	ES_EmergencyStop
)

type ButtonStatus int

const (
	NotActive ButtonStatus = iota
	Pending
	Running
)

// TODO Chat saying this MSG_T_ naming convention is very c -style and noisy in Go
type MessageType int

const (
	MSG_T_StatusReport MessageType = iota

	MSG_T_TaskCreate   // a new task is created/published
	MSG_T_TaskAssign   // a task is assigned to you
	MSG_T_TaskDelegate // a task is assigned to another person
	MSG_T_TaskUpdate   // task changed, Don't think we need it
	MSG_T_TaskComplete // task was completed
	MSG_T_TaskRequest  // someone requests a new task
	MSG_T_LostComs     // A routine to check if you have lost communication
	MSG_T_NewToChannel // Send the latest information
)
