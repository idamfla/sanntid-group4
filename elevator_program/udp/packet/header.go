package packet

type PacketType int

const (
	PKT_T_Heartbeat PacketType = iota
	PKT_T_Data
	PKT_T_MasterData
	PKT_T_BroadcastData
	PKT_T_Ack
	PKT_T_BroadcastAck
	PKT_T_Commit
	PKT_T_BroadcastCommit
	PKT_T_Done
	PKT_T_BroadcastDone
)

type Header struct {
	Seq           uint32
	SessionID     uint32
	PktType       PacketType // Data, Ack, Heartbeat ...
	RecipientAddr string     // where the reply should go
	SenderAddr    string     // where this message came from (IP:Port)
}

func (p PacketType) String() string {
	switch p {
	case PKT_T_Heartbeat:
		return "Heartbeat"
	case PKT_T_Data:
		return "Data"
	case PKT_T_MasterData:
		return "Master Data"
	case PKT_T_BroadcastData:
		return "Broadcast Data"
	case PKT_T_Ack:
		return "Ack"
	case PKT_T_BroadcastAck:
		return "Broadcast Ack"
	case PKT_T_Commit:
		return "Commit"
	case PKT_T_BroadcastCommit:
		return "Broadcast Commit"
	case PKT_T_Done:
		return "Done"
	case PKT_T_BroadcastDone:
		return "Broadcast Done"
	default:
		return "unknown"
	}
}
