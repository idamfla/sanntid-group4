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
	msg string,
) error {

	pck := udp.Packet{
		Header: udp.Header{
			Seq:           seq,
			SessionID:     sessionID,
			MsgType:       msgType,
			RecipientAddr: remoteAddr.String(),
			SenderAddr:    srv.recvConn.LocalAddr().String(),
		},
		Payload: udp.Message{
			Content: msg,
		},
	}

	return udp.SendPacket(srv.sendConn, remoteAddr, pck)
}

func (srv *Server) SendMessage(
	remoteAddr *net.UDPAddr,
	seq uint32,
	sessionID uint32,
	msg string,
) error {

	return srv.send(
		remoteAddr,
		seq,
		sessionID,
		udp.MSG_T_Data,
		msg,
	)
}

func (srv *Server) SendReply(remoteAddr *net.UDPAddr, pck udp.Packet, msgType udp.MessageType) error {
	h := pck.Header

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
		h.Seq+1,
		h.SessionID,
		msgType,
		replyContent,
	)
}

func (srv *Server) SendBroadcast(seq uint32, sessionID uint32, msg string) error {
	// Broadcast address (255.255.255.255/net.IPv4bcast sends to entire subnet) --> we use local subnet
	// localAddr := srv.recvConn.LocalAddr().(*net.UDPAddr)
	// localIP := localAddr.IP

	// broadcastIP := net.IPv4(localIP[0], localIP[1], localIP[2], 255)

	// addr := &net.UDPAddr{
	// 	IP:   net.ParseIP(broadcastIP),
	// 	Port: BroadcastPort,
	// }

	addr := &net.UDPAddr{
		// IP: net.ParseIP("127.0.0.1"),
		IP:   net.ParseIP(HomeBroadcastIP),
		Port: BroadcastPort,
	}

	// TODO make session, send count of recipients

	return srv.SendMessage(addr, seq, sessionID, msg)
}
