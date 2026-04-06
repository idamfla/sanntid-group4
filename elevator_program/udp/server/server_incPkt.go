package server

import (
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

func (srv *Server) deliverToSession(senderAddr *net.UDPAddr, incPkt incomingPacket) {
	sessionID := incPkt.Packet.Header.SessionID
	pktType := incPkt.Packet.Header.PktType

	var ses SessionHandler

	switch pktType {
	case packet.PKT_T_WhoIsAlive:
		ses = srv.handleWhoIsAlive() // TODO need to change ...

	case packet.PKT_T_IAmMaster:
		ses = srv.handleIAmMaster(incPkt) // TODO need to change ...

	case packet.PKT_T_SyncMsg: // TODO this is almost like bsUpdate ... find out how it is different
		ses = srv.handleSyncComplete(senderAddr, incPkt)

	default:
		ses = srv.getOrCreateSession(senderAddr, &sessionID)
	}

	if ses == nil {
		return
	}

	srv.printIncMsg(senderAddr, pktType, incPkt) // TODO DB

	ses.ReceivePacket(incPkt.Packet)
}

func (srv *Server) handleWhoIsAlive() SessionHandler {
	if srv.isSearchingForMaster() {
		return nil
	}
	return srv.getOrCreateMasterElectionSession()
}

func (srv *Server) handleIAmMaster(incPkt incomingPacket) SessionHandler {
	peerKey := incPkt.Packet.Header.SenderAddr
	peer, exists := srv.getPeer(peerKey)

	if !exists || peer == nil {
		fmt.Println("Peer dosent exist") // TODO db
		return nil
	}

	oldMstr := srv.getMasterPeer()

	// already know this master
	if oldMstr != nil && oldMstr.GetAddrString() == peer.GetAddrString() {
		srv.setIsSynced()
		fmt.Println(srv.GetAlias(), "already know master, ignoring") // TODO db
		return nil
	}

	// new master
	if oldMstr != nil {
		oldMstr.ClearMaster()
	}

	peer.SetMaster()
	peer.SetSynced()

	srv.clearIsSynced()

	return srv.getOrCreateMasterElectionSession()
}

func (srv *Server) handleSyncComplete(senderAddr *net.UDPAddr, incPkt incomingPacket) SessionHandler {
	sessionID := incPkt.Packet.Header.SessionID
	recipientAddr := incPkt.Packet.Payload.Addr
	selfAddr := srv.GetRecvString()

	if selfAddr == recipientAddr {
		srv.setIsSynced()
	}

	if peer, exists := srv.getPeer(recipientAddr); exists {
		peer.SetSynced()
	}

	return srv.getOrCreateSession(senderAddr, &sessionID)
}

func (srv *Server) printIncMsg(senderAddr *net.UDPAddr, pktType packet.PacketType, incPkt incomingPacket) {
	if pktType == packet.PKT_T_MasterAck && !srv.IsMaster() {

	} else {
		fmt.Printf(
			`%s, Session %d:
	origin    : %s
	to        : %s
	reply sock: %s
	pktType   : %s
	payload   : %+v
`,
			srv.GetAlias(),
			incPkt.Packet.Header.SessionID,
			incPkt.Packet.Header.Origin.Alias,
			incPkt.Packet.Header.RecipientAddr,
			senderAddr.String(),
			pktType,
			incPkt.Packet.Payload,
		)
	}
}
