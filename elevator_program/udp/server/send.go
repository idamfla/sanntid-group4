package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

func (srv *Server) Send(
	remoteAddr *net.UDPAddr,
	seq uint32,
	sessionID uint32,
	pktType packet.PacketType,
	eMsg message.ElevatorMessage,
) error {

	pkt := packet.Packet{
		Header: packet.Header{
			Seq:           seq,
			SessionID:     sessionID,
			PktType:       pktType,
			RecipientAddr: remoteAddr.String(),
			SenderAddr:    srv.recvConn.LocalAddr().String(),
		},
		Payload: eMsg,
	}

	return packet.SendPacket(srv.sendConn, remoteAddr, pkt)
}

func (srv *Server) startSession(remoteAddr *net.UDPAddr, eMsg message.ElevatorMessage) error {
	if srv.isLocalAddr(remoteAddr) {
		err := fmt.Errorf("Tried to send to oneself %s", remoteAddr.String())
		fmt.Println(err)
		return err
	}

	ses := srv.createSession(remoteAddr, nil)
	ses.QueueSlaveUpdate(eMsg)
	// srv.elevatorTaskQueue()
	return nil
}

// Initiate the broadcast message chain
func (srv *Server) startBroadcast(eMsg message.ElevatorMessage) {
	quorum := srv.getQuorum()
	ses := srv.createBroadcastSession(nil, quorum)

	ses.QueueBroadcastUpdate(eMsg)
}

func (srv *Server) startWhoIsMasterMsg() {
	ses := srv.createBroadcastSession(nil, 0)

	ses.QueueWhoIsMasterMsg()
}

// deciding how to output messages from the server, what type of session they belong to
func (srv *Server) dispatchMessage(outMsg outgoingMessage) {
	defer srv.wg.Done()
	switch outMsg.PktType {
	case packet.PKT_T_SlaveUpdate:
		srv.startSession(outMsg.RemoteAddr, outMsg.EMsg)
	case packet.PKT_T_BroadcastUpdate:
		srv.startBroadcast(outMsg.EMsg)
	case packet.PKT_T_WhoIsMaster:
		srv.startWhoIsMasterMsg()
	}

	// peers := srv.getAliveUnsyncedPeers()
	// for _, peer := range peers {
	// 	peer.QueueMessage(outMsg.Msg)
	// }
}

func (srv *Server) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, eMsg message.ElevatorMessage) {
	pktType := packet.PacketType(protoPktType)
	srv.outgoingMsgCh <- outgoingMessage{
		RemoteAddr: remoteAddr,
		PktType:    pktType,
		EMsg:       eMsg,
	}
}

// --- helper ---
func (srv *Server) isLocalAddr(addr *net.UDPAddr) bool {
	local := srv.recvConn.LocalAddr().(*net.UDPAddr)
	return addr.IP.Equal(local.IP) && addr.Port == local.Port
}
