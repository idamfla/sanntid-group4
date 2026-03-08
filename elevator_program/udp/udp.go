package udp

type MessageType int

const (
	MSG_T_Heartbeat MessageType = iota
	MSG_T_Data
	MSG_T_MasterData
	MSG_T_BroadcastData
	MSG_T_Ack
	MSG_T_BroadcastAck
	MSG_T_Commit
	MSG_T_BroadcastCommit
	MSG_T_Done
	MSG_T_BroadcastDone
)

type Header struct {
	Seq           uint32
	SessionID     uint32
	MsgType       MessageType // Data, Ack, Heartbeat
	RecipientAddr string      // where the reply should go
	SenderAddr    string      // where this message came from (IP:Port)
}

// TODO make sure elevatorStruct has these channels
type ElevatorMessage struct {
	SessionID uint32
	Message   Message
	Done      chan<- Message
}

// in init "ToSessionChans{ToElevator: make(chan Message)..."

func (m MessageType) String() string {
	switch m {
	case MSG_T_Heartbeat:
		return "Heartbeat"
	case MSG_T_Data:
		return "Data"
	case MSG_T_MasterData:
		return "Master Data"
	case MSG_T_BroadcastData:
		return "Broadcast Data"
	case MSG_T_Ack:
		return "Ack"
	case MSG_T_BroadcastAck:
		return "Broadcast Ack"
	case MSG_T_Commit:
		return "Commit"
	case MSG_T_BroadcastCommit:
		return "Broadcast Commit"
	case MSG_T_Done:
		return "Done"
	case MSG_T_BroadcastDone:
		return "Broadcast Done"
	default:
		return "unknown"
	}
}
