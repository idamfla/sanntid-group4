package packet

type PacketType int

const (
	PKT_T_Heartbeat PacketType = iota
	PKT_T_LostConn

	PKT_T_WhoIsMaster
	PKT_T_IAmMaster

	PKT_T_SlaveUpdate
	PKT_T_SlaveUpdateAck

	PKT_T_BroadcastUpdate
	PKT_T_BroadcastAck
	PKT_T_BroadcastCommit
	PKT_T_BroadcastDone

	PKT_T_SyncRequest
	PKT_T_SyncAck

	PKT_T_StateSnapshot
	PKT_T_SnapshotAck

	PKT_T_ElevatorFailed
)

// TODO combine all MasterDoSomethingRequest: slavereport,requestneworder, data. Dont need report and report ack
/*
commit and done is redundant, ack closes one to one msg -> changes only happen when master broadcast. look at msg receiver field for who should do what
- {this is me: ip, who is master} x3 -> (if someone is master) -> wait 5 sec -> one send {i am master, ip}
											|-> sync -> who should be master (should new be master)
*/

type ProtocolPacketType PacketType

const (
	PROTO_PKT_T_Heartbeat       ProtocolPacketType = ProtocolPacketType(PKT_T_Heartbeat) // broadcast
	PROTO_PKT_T_LostConn        ProtocolPacketType = ProtocolPacketType(PKT_T_LostConn)  // broadcast
	PROTO_PKT_T_WhoIsMaster     ProtocolPacketType = ProtocolPacketType(PKT_T_WhoIsMaster)
	PROTO_PKT_T_SlaveUpdate     ProtocolPacketType = ProtocolPacketType(PKT_T_SlaveUpdate)     // slave -> master
	PROTO_PKT_T_BroadcastUpdate ProtocolPacketType = ProtocolPacketType(PKT_T_BroadcastUpdate) // master -> broadcast
	PROTO_PKT_T_SyncRequest     ProtocolPacketType = ProtocolPacketType(PKT_T_SyncRequest)     // slave -> master
	PROTO_PKT_T_StateSnapshot   ProtocolPacketType = ProtocolPacketType(PKT_T_StateSnapshot)   // master -> slave
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
	case PKT_T_LostConn:
		return "Lost connection ..."
	case PKT_T_WhoIsMaster:
		return "Who is master"
	case PKT_T_IAmMaster:
		return "I am master"
	case PKT_T_SlaveUpdate:
		return "Slave Notify"
	case PKT_T_SlaveUpdateAck:
		return "Slave Notify Ack"
	case PKT_T_BroadcastUpdate:
		return "Broadcast Update"
	case PKT_T_BroadcastAck:
		return "Broadcast Ack"
	case PKT_T_BroadcastCommit:
		return "Broadcast Commit"
	case PKT_T_BroadcastDone:
		return "Broadcast Done"
	case PKT_T_SyncRequest:
		return "Sync Request"
	case PKT_T_SyncAck:
		return "Sync Ack"
	case PKT_T_StateSnapshot:
		return "State Snapshot"
	case PKT_T_SnapshotAck:
		return "Snapshot Ack"
	case PKT_T_ElevatorFailed:
		return "Elevator Failed"
	default:
		return "unknown"
	}
}
