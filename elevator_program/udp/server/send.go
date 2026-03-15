package server

import (
	"elevator_program/udp"
	"elevator_program/udp/message"
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

func (srv *Server) Send(
	remoteAddr *net.UDPAddr,
	seq uint32,
	sessionID uint32,
	pktType packet.PacketType,
	msg message.Message,
) error {

	pkt := packet.Packet{
		Header: packet.Header{
			Seq:           seq,
			SessionID:     sessionID,
			PktType:       pktType,
			RecipientAddr: remoteAddr.String(),
			SenderAddr:    srv.recvConn.LocalAddr().String(),
		},
		Payload: msg,
	}

	return packet.SendPacket(srv.sendConn, remoteAddr, pkt)
}

func (srv *Server) startSession(remoteAddr *net.UDPAddr, msg message.Message) {
	localAddr := srv.recvConn.LocalAddr().(*net.UDPAddr)

	// Compare IP and Port
	if remoteAddr.IP.Equal(localAddr.IP) && remoteAddr.Port == localAddr.Port {
		err := fmt.Errorf("Tried to send to oneself %s", remoteAddr.String())
		fmt.Println(err)
		return
	}

	ses := srv.createSession(remoteAddr, nil)
	ses.QueueDataMessage(msg)
}

func (srv *Server) startMasterSession(remoteAddr *net.UDPAddr, msg message.Message) {
	localAddr := srv.recvConn.LocalAddr().(*net.UDPAddr)

	// Compare IP and Port
	if remoteAddr.IP.Equal(localAddr.IP) && remoteAddr.Port == localAddr.Port {
		err := fmt.Errorf("Tried to send to oneself %s", remoteAddr.String())
		fmt.Println(err)
		return
	}

	ses := srv.createSession(remoteAddr, nil)
	ses.QueueMasterMessage(msg)
}

// Initiate the broadcast message chain
func (srv *Server) startBroadcast(msg message.Message) {
	addr := &net.UDPAddr{
		// IP: net.ParseIP("127.0.0.1"),
		IP:   net.ParseIP(udp.HomeBroadcastIP),
		Port: udp.BROADCAST_PORT,
	}

	quorum := srv.getQuorum()
	ses := srv.createBroadcastSession(addr, quorum)

	ses.QueueBroadcastData(msg)
}

func (srv *Server) dispatchMessage(srvMsg outgoingMessage) {
	switch srvMsg.PktType {
	case packet.PKT_T_Data:
		srv.startSession(srvMsg.RemoteAddr, srvMsg.Msg)
	case packet.PKT_T_BroadcastData:
		srv.startBroadcast(srvMsg.Msg)
	case packet.PKT_T_SlaveReport:
		// srv.startMasterSession(srvMsg.RemoteAddr, srvMsg.Msg)
	}
}

func (srv *Server) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, msg message.Message) {
	pktType := packet.PacketType(protoPktType)
	srv.outgoingMsgCh <- outgoingMessage{
		RemoteAddr: remoteAddr,
		PktType:    pktType,
		Msg:        msg,
	}
}
