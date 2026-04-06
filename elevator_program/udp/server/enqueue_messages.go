package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
)

func (srv *Server) QueueWhoIsAliveMsg() {
	srv.queueMessage(srv.GetIdentity(), packet.PROTO_PKT_T_WhoIsAlive, message.ElevatorMessage{})
}

func (srv *Server) QueueIAmMasterMsg() {
	srv.queueMessage(srv.GetIdentity(), packet.PROTO_PKT_T_IAmMaster, message.ElevatorMessage{})
}

func (srv *Server) QueueElectedMasterMsg(masterAddr string) {
	peer, exists := srv.getPeer(masterAddr)
	if !exists {
		fmt.Println("Elected master does not exist ...")
		srv.QueueWhoIsAliveMsg()
		return
	}

	_, addr, _, _, _, _ := peer.Snapshot()

	srv.queueMessage(srv.GetIdentity(), packet.PROTO_PKT_T_ElectedMasterIs, message.ElevatorMessage{Addr: addr.String()})
}

func (srv *Server) QueueRequestTaskExecution(eMsgType message.ElevatorMessageType) {
	srv.queueMessage(
		srv.GetIdentity(),
		packet.PROTO_PKT_T_RequestTaskExecution,
		message.ElevatorMessage{
			EMsgType: eMsgType,
			Addr:     srv.GetRecvString(),
		},
	)
}

func (srv *Server) QueueSyncMsg(eMsg message.ElevatorMessage) {
	origin, exists := srv.resolveOrigin(eMsg)
	if !exists {
		fmt.Println("The origin of this message is unknown ...")
		srv.QueueWhoIsAliveMsg()
		return
	}

	srv.queueMessage(origin, packet.PROTO_PKT_T_SyncMsg, eMsg)
}

func (srv *Server) resolveOrigin(eMsg message.ElevatorMessage) (origin packet.Identity, exists bool) {
	key := eMsg.Addr

	if key == srv.GetRecvString() {
		return srv.GetIdentity(), true
	}

	peer, exists := srv.getPeer(key)
	srv.PrintPeers()
	if !exists {
		<-srv.stop
		return packet.Identity{}, false
	}

	alias, addr, _, _, _, _ := peer.Snapshot()

	return packet.Identity{
			Identifier: addr.String(),
			Alias:      alias,
		},
		true
}

// TODO this will be private in the end, it's more of a helper
func (srv *Server) queueMessage(origin packet.Identity, protoPktType packet.ProtocolPacketType, eMsg message.ElevatorMessage) {
	pktType := packet.PacketType(protoPktType)

	select {
	case srv.outgoingMsgCh <- packet.OutgoingMessage{
		Origin:  origin,
		PktType: pktType,
		EMsg:    eMsg,
	}:
	default:
		fmt.Println("Can't queue message, servers messageQueue is full")
	}
}
