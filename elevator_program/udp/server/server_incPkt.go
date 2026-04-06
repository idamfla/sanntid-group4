package server

import (
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

func (srv *Server) deliverToSession(senderAddr *net.UDPAddr, pkt packet.Packet) {
	sessionID := pkt.Header.SessionID
	pktType := pkt.Header.PktType

	var ses SessionHandler

	switch pktType {
	case packet.PKT_T_WhoIsAlive:
		ses = srv.handleWhoIsAlive() // TODO need to change ...

	case packet.PKT_T_IAmMaster:
		ses = srv.handleIAmMaster(pkt) // TODO need to change ...

	case packet.PKT_T_SyncComplete:
		ses = srv.handleSyncComplete(senderAddr, pkt)

	default:
		ses = srv.getOrCreateSession(senderAddr, &sessionID)
	}

	if ses == nil {
		return
	}

	srv.printIncMsg(senderAddr, pktType, pkt) // TODO DB

	ses.ReceivePacket(pkt)
}

func (srv *Server) handleWhoIsAlive() SessionHandler {
	if srv.isSearchingForMaster() {
		return nil
	}
	return srv.getOrCreateMasterElectionSession()
}

func (srv *Server) handleIAmMaster(pkt packet.Packet) SessionHandler {
	peerKey := pkt.Header.SenderAddr
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

func (srv *Server) handleSyncComplete(senderAddr *net.UDPAddr, pkt packet.Packet) SessionHandler {
	sessionID := pkt.Header.SessionID
	recipientAddr := pkt.Header.Origin.Identifier

	fmt.Println("SYNC COMP", recipientAddr)

	if srv.GetRecvString() == recipientAddr {
		srv.setIsSynced()
	}

	if peer, exists := srv.getPeer(recipientAddr); exists {
		peer.SetSynced()
	}

	return srv.getOrCreateSession(senderAddr, &sessionID)
}

func (srv *Server) printIncMsg(senderAddr *net.UDPAddr, pktType packet.PacketType, pkt packet.Packet) {
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
			pkt.Header.SessionID,
			pkt.Header.Origin.Alias,
			pkt.Header.RecipientAddr,
			senderAddr.String(),
			pktType,
			pkt.Payload,
		)
	}
}
