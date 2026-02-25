package udp

type MessageType int

const (
	MSG_T_Heartbeat MessageType = iota
	MSG_T_Data
	MSG_T_Ack
	MSG_T_Commit
	MSG_T_Done
)

type Header struct {
	Seq       uint32
	SessionID uint32
	MsgType   MessageType // Data, Ack, Heartbeat
	ReplyIP   string
	ReplyPort int
}
