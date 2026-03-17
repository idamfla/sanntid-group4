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

func (srv *Server) startSession(remoteAddr *net.UDPAddr, msg message.Message) error {
	if srv.isLocalAddr(remoteAddr) {
		err := fmt.Errorf("Tried to send to oneself %s", remoteAddr.String())
		fmt.Println(err)
		return err
	}

	ses := srv.createSession(remoteAddr, nil)
	ses.QueueDataMessage(msg)
	// srv.elevatorTaskQueue()
	return nil
}

func (srv *Server) startReport(remoteAddr *net.UDPAddr, msg message.Message) error {
	if srv.isLocalAddr(remoteAddr) {
		err := fmt.Errorf("Tried to report to oneself %s", remoteAddr.String())
		fmt.Println(err)
		return err
	}

	ses := srv.createSession(remoteAddr, nil)
	ses.QueueMasterMessage(msg)
	return nil
}

// Initiate the broadcast message chain
func (srv *Server) startBroadcast(msg message.Message) {
	quorum := srv.getQuorum()
	ses := srv.createBroadcastSession(srv.broadcastAddr, quorum)

	ses.QueueBroadcastUpdate(msg)
}

func (srv *Server) startStateSync() {
	fmt.Printf("Server %s: about to request state sync\n", srv.ID)
	ses := srv.createSession(srv.broadcastAddr, nil)
	ses.QueueStateSync()
}

func (srv *Server) dispatchMessage(outMsg outgoingMessage) {
	defer srv.wg.Done()
	switch outMsg.PktType {
	case packet.PKT_T_Data:
		srv.startSession(outMsg.RemoteAddr, outMsg.Msg)
	case packet.PKT_T_BroadcastUpdate:
		srv.startBroadcast(outMsg.Msg)
	case packet.PKT_T_SlaveReport:
		// srv.startMasterSession(srvMsg.RemoteAddr, srvMsg.Msg)
	case packet.PKT_T_StateSync:
		srv.startStateSync()
		// TODO what to do when you are completely new???
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

// --- helper ---
func (srv *Server) isLocalAddr(addr *net.UDPAddr) bool {
	local := srv.recvConn.LocalAddr().(*net.UDPAddr)
	return addr.IP.Equal(local.IP) && addr.Port == local.Port
}
