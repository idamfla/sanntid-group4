package message

type FaultType int

const (
	FAULT_T_LostConn FaultType = iota
	FAULT_T_LostMaster
	FAULT_T_ElevatorFailed
	FAULT_T_BroadcastFailedToRespond
	FAULT_T_TaskRunningErr
)

type FaultMessage struct {
	ID        string
	FaultType FaultType
	IsMaster  bool

	SelfAddr string
	Peers    []string
}
