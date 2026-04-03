package packet

type PacketType int

const (
	PKT_T_Heartbeat PacketType = iota
	PKT_T_LostConn

	PKT_T_WhoIsAlive
	PKT_T_IAmAlive
	PKT_T_IAmMaster // respond with MasterAck, if mstr not known

	PKT_T_ElectedMasterIs // if this is you, respond with IAmMaster
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

	PKT_T_ElevatorFailed // todo do i need this?
)

// TODO combine all MasterDoSomethingRequest: slavereport,requestneworder, data. Dont need report and report ack
/*
commit and done is redundant, ack closes one to one msg -> changes only happen when master broadcast. look at msg receiver field for who should do what
- {this is me: ip, who is master} x3 -> (if someone is master) -> wait 5 sec -> one send {i am master, ip}
											|-> sync -> who should be master (should new be master)
*/

type ProtocolPacketType PacketType

const (
	PROTO_PKT_T_Heartbeat            ProtocolPacketType = ProtocolPacketType(PKT_T_Heartbeat)            // broadcast
	PROTO_PKT_T_LostConn             ProtocolPacketType = ProtocolPacketType(PKT_T_LostConn)             // broadcast
	PROTO_PKT_T_WhoIsAlive           ProtocolPacketType = ProtocolPacketType(PKT_T_WhoIsAlive)           //broadcast TODO this should be automatic ...
	PROTO_PKT_T_Snapshot             ProtocolPacketType = ProtocolPacketType(PKT_T_Snapshot)             // master -> slave
	PROTO_PKT_T_CatchupUpdate        ProtocolPacketType = ProtocolPacketType(PKT_T_CatchupUpdate)        // master -> slave
	PROTO_PKT_T_SyncComplete         ProtocolPacketType = ProtocolPacketType(PKT_T_SyncComplete)         // master -> slave
	PROTO_PKT_T_BroadcastUpdate      ProtocolPacketType = ProtocolPacketType(PKT_T_BroadcastUpdate)      // master -> broadcast
	PROTO_PKT_T_SlaveUpdate          ProtocolPacketType = ProtocolPacketType(PKT_T_SlaveUpdate)          // slave -> master
	PROTO_PKT_T_RequestTaskExecution ProtocolPacketType = ProtocolPacketType(PKT_T_RequestTaskExecution) // slave -> master
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
