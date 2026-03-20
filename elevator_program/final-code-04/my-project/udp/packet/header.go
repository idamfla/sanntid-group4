package packet

type PacketType int

const (
	PKT_T_Heartbeat PacketType = iota
	PKT_T_IAmAlive
	PKT_T_LostConn

	PKT_T_WhoIsMaster
	PKT_T_IAmMaster
	PKT_T_MasterAck

	PKT_T_Snapshot
	PKT_T_SnapshotAck

	PKT_T_CatchupUpdate
	PKT_T_CatchupAck

	PKT_T_SyncComplete
	PKT_T_SyncAck
	PKT_T_SyncCommit
	PKT_T_SyncDone

	PKT_T_BroadcastUpdate
	PKT_T_BroadcastAck
	PKT_T_BroadcastCommit
	PKT_T_BroadcastDone

	PKT_T_SlaveUpdate
	PKT_T_SlaveUpdateAck

	PKT_T_RequestTaskExecution
	PKT_T_RequestTaskExecutionAck

	PKT_T_ElevatorFailed
)

type ProtocolPacketType PacketType

const (
	PROTO_PKT_T_Heartbeat            ProtocolPacketType = ProtocolPacketType(PKT_T_Heartbeat)
	PROTO_PKT_T_LostConn             ProtocolPacketType = ProtocolPacketType(PKT_T_LostConn)
	PROTO_PKT_T_WhoIsMaster          ProtocolPacketType = ProtocolPacketType(PKT_T_WhoIsMaster)
	PROTO_PKT_T_Snapshot             ProtocolPacketType = ProtocolPacketType(PKT_T_Snapshot)
	PROTO_PKT_T_CatchupUpdate        ProtocolPacketType = ProtocolPacketType(PKT_T_CatchupUpdate)
	PROTO_PKT_T_SyncComplete         ProtocolPacketType = ProtocolPacketType(PKT_T_SyncComplete)
	PROTO_PKT_T_BroadcastUpdate      ProtocolPacketType = ProtocolPacketType(PKT_T_BroadcastUpdate)
	PROTO_PKT_T_SlaveUpdate          ProtocolPacketType = ProtocolPacketType(PKT_T_SlaveUpdate)
	PROTO_PKT_T_RequestTaskExecution ProtocolPacketType = ProtocolPacketType(PKT_T_RequestTaskExecution)
)

type Header struct {
	Seq           uint32
	SessionID     uint32
	PktType       PacketType
	RecipientAddr string
	SenderAddr    string
}

func (p PacketType) String() string {
	switch p {
	case PKT_T_Heartbeat:
		return "Heartbeat"
	case PKT_T_LostConn:
		return "Lost connection ..."

	case PKT_T_IAmAlive:
		return "I am alive"

	case PKT_T_WhoIsMaster:
		return "Who is master"
	case PKT_T_IAmMaster:
		return "I am master"
	case PKT_T_MasterAck:
		return "Master Ack"

	case PKT_T_Snapshot:
		return "Snapshot"
	case PKT_T_SnapshotAck:
		return "Snapshot Ack"

	case PKT_T_CatchupUpdate:
		return "Catch Up Update"
	case PKT_T_CatchupAck:
		return "Catch Up Ack"

	case PKT_T_SyncComplete:
		return "Synchronization Complete"
	case PKT_T_SyncAck:
		return "Synchronization Ack"
	case PKT_T_SyncCommit:
		return "Synchronization Commit"
	case PKT_T_SyncDone:
		return "Synchronization Done"

	case PKT_T_BroadcastUpdate:
		return "Broadcast Update"
	case PKT_T_BroadcastAck:
		return "Broadcast Ack"
	case PKT_T_BroadcastCommit:
		return "Broadcast Commit"
	case PKT_T_BroadcastDone:
		return "Broadcast Done"

	case PKT_T_SlaveUpdate:
		return "Slave Update"
	case PKT_T_SlaveUpdateAck:
		return "Slave Update Ack"

	case PKT_T_RequestTaskExecution:
		return "Request Task Execution"
	case PKT_T_RequestTaskExecutionAck:
		return "Request Task Execution Ack"

	case PKT_T_ElevatorFailed:
		return "Elevator Failed"
	default:
		return "unknown"
	}
}
