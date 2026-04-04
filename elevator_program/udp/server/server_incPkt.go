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

	case packet.PKT_T_SyncComplete:
		ses = srv.handleSyncComplete(senderAddr, incPkt)

	default:
		if packet.IsBroadcastPkt(pktType) && srv.isSynced() == false {
			fmt.Println(srv.ID, "is not synced so it can take no new updates") // TODO db
			return
		}

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
	srv.mu.Lock()

	peerID := incPkt.Packet.Header.SenderAddr
	peer, exists := srv.peers[peerID]

	if !exists || peer == nil {
		srv.mu.Unlock()
		fmt.Println("Peer dosent exist") // TODO db
		return nil
	}

	oldMstr := srv.getMasterPeerUnsafe()

	// already know this master
	if oldMstr != nil && oldMstr.Addr.String() == peer.Addr.String() {
		srv.SetIsSynced(true)
		srv.mu.Unlock()
		fmt.Println(srv.ID, "already know master, ignoring") // TODO db
		return nil
	}

	// new master
	if oldMstr != nil {
		oldMstr.SetMaster(false)
	}

	peer.SetMaster(true)
	peer.SetIsSynced(true)

	srv.SetIsSynced(false)

	srv.mu.Unlock()

	srv.queueSyncRequest()
	return srv.getOrCreateMasterElectionSession()
}

func (srv *Server) handleSyncComplete(senderAddr *net.UDPAddr, incPkt incomingPacket) SessionHandler {
	srv.mu.Lock()
	sessionID := incPkt.Packet.Header.SessionID
	recipientAddr := incPkt.Packet.Payload.Addr
	selfAddr := srv.GetRecvString()

	if selfAddr == recipientAddr {
		srv.SetIsSynced(true)
	}

	if peer, exists := srv.peers[recipientAddr]; exists {
		peer.SetIsSynced(true)
	}
	srv.mu.Unlock()

	return srv.getOrCreateSession(senderAddr, &sessionID)
}

func (srv *Server) printIncMsg(senderAddr *net.UDPAddr, pktType packet.PacketType, incPkt incomingPacket) {
	if pktType == packet.PKT_T_MasterAck && !srv.IsMaster() {

	} else {
		fmt.Printf(
			`%s, Session %d:
	sent from : %s
	to        : %s
	reply sock: %s
	pktType   : %s
	payload   : %+v
`,
			srv.ID,
			incPkt.Packet.Header.SessionID,
			incPkt.Addr.String(),
			incPkt.Packet.Header.RecipientAddr,
			senderAddr.String(),
			pktType,
			incPkt.Packet.Payload,
		)
	}
}
