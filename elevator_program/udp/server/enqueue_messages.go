package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
)

func (srv *Server) QueueWhoIsAliveMsg() {
	srv.QueueMessage(packet.PROTO_PKT_T_WhoIsAlive, message.ElevatorMessage{})
}

func (srv *Server) QueueIAmMasterMsg() {
	srv.QueueMessage(packet.PROTO_PKT_T_IAmMaster, message.ElevatorMessage{})
}

func (srv *Server) QueueElectedMasterMsg(masterAddr string) {
	peer, exists := srv.getPeer(masterAddr)
	if !exists {
		fmt.Println("Elected master does not exist ...")
		srv.QueueWhoIsAliveMsg()
		return
	}

	_, addr, _, _, _, _ := peer.Snapshot()

	srv.QueueMessage(packet.PROTO_PKT_T_ElectedMasterIs, message.ElevatorMessage{Addr: addr.String()})
}

// TODO this will be private in the end, it's more of a helper
func (srv *Server) QueueMessage(protoPktType packet.ProtocolPacketType, eMsg message.ElevatorMessage) {
	pktType := packet.PacketType(protoPktType)

	select {
	case srv.outgoingMsgCh <- packet.OutgoingMessage{
		Origin: packet.Identity{
			Identifier: srv.GetRecvString(),
			Alias:      srv.GetAlias(),
		},
		PktType: pktType,
		EMsg:    eMsg,
	}:
	default:
		fmt.Println("Can't queue message, servers messageQueue is full")
	}
}
