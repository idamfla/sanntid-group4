package server

import (
	"elevator_program/udp"
	"net"
)

// TODO dont send string but rather the Message-struct
func (
	srv *Server) send(
	remoteAddr *net.UDPAddr,
	seq uint32,
	sessionID uint32,
	msgType udp.MessageType,
	senderAddr string,
	msg string,
) error {

	pck := udp.Packet{
		Header: udp.Header{
			Seq:           seq,
			SessionID:     sessionID,
			MsgType:       msgType,
			SenderAddr:    senderAddr,
			RecipientAddr: remoteAddr.String(),
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

	localAddr := srv.recvConn.LocalAddr().(*net.UDPAddr)

	return srv.send(
		remoteAddr,
		seq,
		sessionID,
		udp.MSG_T_Data,
		localAddr.String(),
		msg,
	)
}

func (srv *Server) SendReply(remoteAddr *net.UDPAddr, pck udp.Packet, msgType udp.MessageType) error {
	h := pck.Header

	// TODO maybe it's own function, what to when skipping commit messages and go straight to "done"
	replyContent := ""
	switch msgType {
	case udp.MSG_T_Ack:
		replyContent = "ACK"
	case udp.MSG_T_Commit:
		replyContent = "COMMIT"
	case udp.MSG_T_Done:
		replyContent = "DONE"
	}

	return srv.send(
		remoteAddr,
		h.Seq+1,
		h.SessionID,
		msgType,
		h.RecipientAddr,
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
		IP:   net.ParseIP(Group4IP),
		Port: BroadcastPort,
	}

	// TODO make session, send count of recipients

	return srv.SendMessage(addr, seq, sessionID, msg)
}
