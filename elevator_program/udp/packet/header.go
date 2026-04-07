package packet

type PacketType int

const (
	PKT_T_Heartbeat PacketType = iota
	PKT_T_LostConn

	PKT_T_WhoIsAlive // initiater
	PKT_T_IAmAlive
	PKT_T_IAmMaster // initiater

	PKT_T_ElectedMasterIs // if this is you, respond with IAmMaster
	PKT_T_MasterAck

	PKT_T_SyncMsg // initiater
	PKT_T_SyncMsgAck
	PKT_T_SyncMsgCommit
	PKT_T_SyncComplete

	PKT_T_BroadcastUpdate // initiater
	PKT_T_BroadcastAck
	PKT_T_BroadcastCommit
	PKT_T_BroadcastDone

	PKT_T_RequestTaskExecution // initiater
	PKT_T_RequestTaskExecutionAck

	PKT_T_ElevatorFailed
)

type ProtocolPacketType PacketType

const (
	PROTO_PKT_T_Heartbeat            ProtocolPacketType = ProtocolPacketType(PKT_T_Heartbeat)       // broadcast
	PROTO_PKT_T_LostConn             ProtocolPacketType = ProtocolPacketType(PKT_T_LostConn)        // broadcast
	PROTO_PKT_T_WhoIsAlive           ProtocolPacketType = ProtocolPacketType(PKT_T_WhoIsAlive)      //broadcast TODO this should be automatic ...
	PROTO_PKT_T_IAmMaster            ProtocolPacketType = ProtocolPacketType(PKT_T_IAmMaster)       //broadcast TODO this should be automatic ...
	PROTO_PKT_T_ElectedMasterIs      ProtocolPacketType = ProtocolPacketType(PKT_T_ElectedMasterIs) //broadcast TODO this should be automatic ...
	PROTO_PKT_T_SyncMsg              ProtocolPacketType = ProtocolPacketType(PKT_T_SyncMsg)         // master -> slave
	PROTO_PKT_T_SyncComplete         ProtocolPacketType = ProtocolPacketType(PKT_T_SyncComplete)
	PROTO_PKT_T_BroadcastUpdate      ProtocolPacketType = ProtocolPacketType(PKT_T_BroadcastUpdate)      // master -> broadcast
	PROTO_PKT_T_RequestTaskExecution ProtocolPacketType = ProtocolPacketType(PKT_T_RequestTaskExecution) // slave -> master
)

type Header struct {
	Origin        Identity
	Seq           uint32
	PktType       PacketType // Data, Ack, Heartbeat ...
	RecipientAddr string     // where the reply should go
	SenderAddr    string     // where this message came from (IP:Port)
}

type Identity struct {
	ID         uint32
	Identifier string
	Alias      string
}

func (p PacketType) String() string {
	switch p {
	case PKT_T_Heartbeat:
		return "Heartbeat"
	case PKT_T_LostConn:
		return "Lost connection ..."

	case PKT_T_WhoIsAlive:
		return "Who is alive"
	case PKT_T_IAmAlive:
		return "I am alive"
	case PKT_T_ElectedMasterIs:
		return "Elected master is"
	case PKT_T_IAmMaster:
		return "I am master"
	case PKT_T_MasterAck:
		return "Master Ack"

	case PKT_T_SyncMsg:
		return "Synchronization Message"
	case PKT_T_SyncMsgAck:
		return "Synchronization Message Ack"
	case PKT_T_SyncMsgCommit:
		return "Synchronization Message Commit"
	case PKT_T_SyncComplete:
		return "Synchronization Complete"

	case PKT_T_BroadcastUpdate:
		return "Broadcast Update"
	case PKT_T_BroadcastAck:
		return "Broadcast Ack"
	case PKT_T_BroadcastCommit:
		return "Broadcast Commit"
	case PKT_T_BroadcastDone:
		return "Broadcast Done"

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
