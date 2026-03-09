package server

import (
	"elevator_program/udp/message"
	"elevator_program/udp/packet"
	"net"
)

// TODO dont send string but rather the Message-struct
func (srv *Server) send(
	remoteAddr *net.UDPAddr,
	seq uint32,
	sessionID uint32,
	msgType packet.PacketType,
	msg message.Message,
) error {

	pkt := packet.Packet{
		Header: packet.Header{
			Seq:           seq,
			SessionID:     sessionID,
			PktType:       msgType,
			RecipientAddr: remoteAddr.String(),
			SenderAddr:    srv.recvConn.LocalAddr().String(),
		},
		Payload: msg,
	}

	return packet.SendPacket(srv.sendConn, remoteAddr, pkt)
}

func (srv *Server) SendMessage(
	remoteAddr *net.UDPAddr,
	seq uint32,
	sessionID uint32,
	msg message.Message,
) error {

	return srv.send(
		remoteAddr,
		seq,
		sessionID,
		packet.PKT_T_Data,
		msg,
	)
}

func (srv *Server) SendReply(remoteAddr *net.UDPAddr, seq uint32, sessionID uint32, msgType packet.PacketType) error {
	// TODO maybe it's own function, what to when skipping commit messages and go straight to "done"
	replyContent := ""
	switch msgType {
	case packet.PKT_T_Ack:
		replyContent = srv.ID + " received: ACK"
	case packet.PKT_T_Commit:
		replyContent = srv.ID + " received: COMMIT"
	case packet.PKT_T_Done:
		replyContent = srv.ID + " received: DONE"
	}

	return srv.send(
		remoteAddr,
		seq,
		sessionID,
		msgType,
		message.Message{Content: replyContent},
	)
}

func (srv *Server) SendBroadcast(seq uint32, sessionID uint32, msg message.Message) error {
	addr := &net.UDPAddr{
		// IP: net.ParseIP("127.0.0.1"),
		IP:   net.ParseIP(HomeBroadcastIP),
		Port: BroadcastPort,
	}

	return srv.send(
		addr,
		seq,
		sessionID,
		packet.PKT_T_BroadcastData,
		msg,
	)
}
