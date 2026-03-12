package server

import (
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

func (srv *Server) StartSession(remoteAddr *net.UDPAddr, msg message.Message) {
	localAddr := srv.recvConn.LocalAddr().(*net.UDPAddr)

	// Compare IP and Port
	if remoteAddr.IP.Equal(localAddr.IP) && remoteAddr.Port == localAddr.Port {
		err := fmt.Errorf("Tried to send to oneself %s", remoteAddr.String())
		fmt.Println(err)
		return
	}

	ses := srv.createSession(remoteAddr)
	ses.SendDataMessage(msg)
}

func (srv *Server) StartMasterSession(remoteAddr *net.UDPAddr, msg message.Message) {
	localAddr := srv.recvConn.LocalAddr().(*net.UDPAddr)

	// Compare IP and Port
	if remoteAddr.IP.Equal(localAddr.IP) && remoteAddr.Port == localAddr.Port {
		err := fmt.Errorf("Tried to send to oneself %s", remoteAddr.String())
		fmt.Println(err)
		return
	}

	ses := srv.createSession(remoteAddr)
	ses.SendMasterMessage(msg)
}

// Initiate the broadcast message chain
func (srv *Server) StartBroadcast(msg message.Message) {
	addr := &net.UDPAddr{
		// IP: net.ParseIP("127.0.0.1"),
		IP:   net.ParseIP(HomeBroadcastIP),
		Port: BroadcastPort,
	}

	ses := srv.createSession(addr)

	ses.SendBroadcastData(msg)
}
