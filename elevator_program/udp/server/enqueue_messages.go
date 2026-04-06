package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

func (srv *Server) QueueWhoIsAliveMsg() {
	srv.QueueMessage(nil, packet.PROTO_PKT_T_WhoIsAlive, message.ElevatorMessage{})
}

func (srv *Server) QueueIAmMasterMsg() {
	fmt.Println("queue i am master msg")
	srv.QueueMessage(nil, packet.PROTO_PKT_T_IAmMaster, message.ElevatorMessage{})
}

func (srv *Server) QueueElectedMasterMsg(masterAddr string) {
	peer, exists := srv.getPeer(masterAddr)
	if !exists {
		fmt.Println("Elected master does not exist ...")
		srv.QueueWhoIsAliveMsg()
		return
	}

	id, addr, _, _, _, _ := peer.Snapshot()

	srv.QueueMessage(nil, packet.PROTO_PKT_T_ElectedMasterIs, message.ElevatorMessage{ID: id, Addr: addr.String()})
}

func (srv *Server) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, eMsg message.ElevatorMessage) {
	pktType := packet.PacketType(protoPktType)

	select {
	case srv.outgoingMsgCh <- packet.OutgoingMessage{
		Origin: packet.Identity{
			Identifier: srv.GetRecvString(),
			Alias:      srv.GetAlias(),
		},
		RemoteAddr: remoteAddr,
		PktType:    pktType,
		EMsg:       eMsg,
	}:
	default:
		fmt.Println("Can't queue message, servers messageQueue is full")
	}
}
