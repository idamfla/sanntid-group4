package packet

type PacketType int

const (
	PKT_T_Heartbeat PacketType = iota
	PKT_T_LostConn
	PKT_T_Data
	PKT_T_SlaveReport
	PKT_T_RequestNewOrder
	PKT_T_BroadcastUpdate
	PKT_T_StateSync
	PKT_T_Ack
	PKT_T_BroadcastUpdateAck
	PKT_T_ReportAck
	PKT_T_Commit
	PKT_T_ElevatorFailed
	PKT_T_BroadcastCommit
	PKT_T_Done
	PKT_T_BroadcastDone
)

/*
Heartbeat
LostConn
Notify
NotifyAck
BroadcastData
BroadcastAck
BroadcastCommit
BroadcastDone
ElevatorFailed
*/

type ProtocolPacketType PacketType

const (
	PROTO_PKT_T_Heartbeat       ProtocolPacketType = ProtocolPacketType(PKT_T_Heartbeat)       // broadcast
	PROTO_PKT_T_LostConn        ProtocolPacketType = ProtocolPacketType(PKT_T_LostConn)        // broadcast
	PROTO_PKT_T_Data            ProtocolPacketType = ProtocolPacketType(PKT_T_Data)            // master -> slave
	PROTO_PKT_T_SlaveReport     ProtocolPacketType = ProtocolPacketType(PKT_T_SlaveReport)     // slave -> master
	PROTO_PKT_T_RequestNewOrder ProtocolPacketType = ProtocolPacketType(PKT_T_RequestNewOrder) // slave -> master
	PROTO_PKT_T_BroadcastUpdate ProtocolPacketType = ProtocolPacketType(PKT_T_BroadcastUpdate) // master -> broadcast
	PROTO_PKT_T_StateSync       ProtocolPacketType = ProtocolPacketType(PKT_T_StateSync)       // unknown -> broadcast
)

/*
Heartbeat: broadcast: data -> ack
LostConn: slave broadcast: data -> ack
NotifyMaster: slave -> master: data -> ack
UpdateSystem: master broadcast: data -> ack -> ack counter -> commit -> commit self -> done
NewToChannel: unknown broadcast: data -> ack
*/

// TODO
/*
1. lost conn - broadcast
2. master control - ask for new order, one-to-one
3. button press - have master broadcast, one-to-one
4. master assign task - one-to-one
5. master broadcast - broadcast
6. new to network - get id etc ..., boadcast
7. heartbeat - master broadcast
*/

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
	case PKT_T_LostConn:
		return "Lost connection ..."
	case PKT_T_Data:
		return "Data"
	case PKT_T_SlaveReport:
		return "Slave Report"
	case PKT_T_RequestNewOrder:
		return "Slave requested new order"
	case PKT_T_BroadcastUpdate:
		return "Broadcast Update"
	case PKT_T_StateSync:
		return "New Node"
	case PKT_T_Ack:
		return "Ack"
	case PKT_T_BroadcastUpdateAck:
		return "Broadcast Update Ack"
	case PKT_T_ReportAck:
		return "Master Ack"
	case PKT_T_Commit:
		return "Commit"
	case PKT_T_ElevatorFailed:
		return "Elevator Failed"
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
