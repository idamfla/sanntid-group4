package server

import (
	"elevator_program/udp"
	"net"
)

// TODO dont send string but rather the Message-struct
func (srv *Server) send(
	remoteAddr *net.UDPAddr,
	seq uint32,
	sessionID uint32,
	msgType udp.MessageType,
	msg udp.Message,
) error {

	pck := udp.Packet{
		Header: udp.Header{
			Seq:           seq,
			SessionID:     sessionID,
			MsgType:       msgType,
			RecipientAddr: remoteAddr.String(),
			SenderAddr:    srv.recvConn.LocalAddr().String(),
		},
		Payload: msg,
	}

	return udp.SendPacket(srv.sendConn, remoteAddr, pck)
}

func (srv *Server) SendMessage(
	remoteAddr *net.UDPAddr,
	seq uint32,
	sessionID uint32,
	msg udp.Message,
) error {

	return srv.send(
		remoteAddr,
		seq,
		sessionID,
		udp.MSG_T_Data,
		msg,
	)
}

func (srv *Server) SendReply(remoteAddr *net.UDPAddr, seq uint32, sessionID uint32, msgType udp.MessageType) error {
	// TODO maybe it's own function, what to when skipping commit messages and go straight to "done"
	replyContent := ""
	switch msgType {
	case udp.MSG_T_Ack:
		replyContent = srv.ID + " received: ACK"
	case udp.MSG_T_Commit:
		replyContent = srv.ID + " received: COMMIT"
	case udp.MSG_T_Done:
		replyContent = srv.ID + " received: DONE"
	}

	return srv.send(
		remoteAddr,
		seq,
		sessionID,
		msgType,
		udp.Message{Content: replyContent},
	)
}

func (srv *Server) SendBroadcast(seq uint32, sessionID uint32, msg udp.Message) error {
	addr := &net.UDPAddr{
		// IP: net.ParseIP("127.0.0.1"),
		IP:   net.ParseIP(HomeBroadcastIP),
		Port: BroadcastPort,
	}

	return srv.send(
		addr,
		seq,
		sessionID,
		udp.MSG_T_BroadcastData,
		msg,
	)
}
