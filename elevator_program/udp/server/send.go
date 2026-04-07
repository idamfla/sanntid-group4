package server

import (
	"elevator_program/udp/packet"
	"elevator_program/udp/session"
	"fmt"
	"net"
)

func (srv *Server) Send(
	ses *session.Session,
	pktType packet.PacketType,
	outMsg packet.OutgoingMessage,
) error {
	senderAddr := srv.GetRecvString()

	data, err := ses.GenerateDataPacket(senderAddr, pktType, outMsg)
	if err != nil {
		fmt.Println("Encode error:", err)
		return err
	}

	_, err = srv.getSendConn().WriteToUDP(data, ses.GetPeerAddr())
	if err != nil {
		fmt.Println("Send error:", err)
		return err
	}

	return nil
}

// deciding how to output messages from the server, what type of session should start
func (srv *Server) handleOutMsg(outMsg packet.OutgoingMessage) {
	defer srv.WgDone()
	switch outMsg.PktType {
	case packet.PKT_T_WhoIsAlive, packet.PKT_T_IAmMaster, packet.PKT_T_ElectedMasterIs:
		srv.dispatchMasterElectionMsg(outMsg)

	case packet.PKT_T_BroadcastUpdate, packet.PKT_T_SyncMsg:
		srv.dispatchBroadcastMsg(outMsg)

	case packet.PKT_T_SyncComplete:
		srv.dispatchSyncComplete(outMsg)

	case packet.PKT_T_RequestTaskExecution:
		srv.dispatchToMasterMsg(outMsg)
	}
}

func (srv *Server) dispatchMasterElectionMsg(outMsg packet.OutgoingMessage) {
	ws := srv.getOrCreateMasterElectionSession()
	if ws == nil {
		return
	}

	switch outMsg.PktType {
	case packet.PKT_T_IAmMaster:
		if oldMstr := srv.getMasterPeer(); oldMstr != nil {
			oldMstr.ClearMaster()
		}

		ready := make(chan struct{}, 1)
		ready <- struct{}{}
		srv.QueueElevatorTask(outMsg.EMsg, nil, ready)

		srv.setSelfAsMaster()
		srv.setSynced()
	case packet.PKT_T_WhoIsAlive:
		srv.clearSelfAsMaster()
		srv.clearAllAlive()
	}

	ws.QueueDirectMsg(outMsg.PktType, outMsg)
}

func (srv *Server) dispatchBroadcastMsg(outMsg packet.OutgoingMessage) {
	if !srv.IsMaster() {
		fmt.Println(srv.GetAlias(), "is not master, can't broadcast like one ...")
		return
	}

	srv.startStateBroadcast(outMsg)
}

func (srv *Server) dispatchSyncComplete(outMsg packet.OutgoingMessage) {
	ses, exists := srv.getSession(outMsg.Origin.ID)
	if !exists {
		fmt.Printf("Tried to dispatch a %s to a session that do not exist ...\n", outMsg.PktType)
		return
	}

	srv.updateSyncFromMsg(outMsg)
	ses.QueueDirectMsg(outMsg.PktType, outMsg)
}

func (srv *Server) dispatchToMasterMsg(outMsg packet.OutgoingMessage) {
	mstr := srv.getMasterPeer()
	if mstr == nil {
		fmt.Println(srv.GetAlias(), "dosen't know who master is")
		srv.QueueWhoIsAliveMsg()
		return
	}

	srv.startSession(mstr.Addr, outMsg)
}

// --- start sessions ---

func (srv *Server) startSession(remoteAddr *net.UDPAddr, outMsg packet.OutgoingMessage) {
	ses := srv.getOrCreateSession(remoteAddr, nil)
	if ses == nil {
		fmt.Printf("Server %s: could not start session, failed to getOrCreate ...\n", srv.GetAlias())
		return
	}
	ses.QueueDirectMsg(outMsg.PktType, outMsg)
}

// Initiate the broadcast message chain
func (srv *Server) startStateBroadcast(outMsg packet.OutgoingMessage) {
	quorum := srv.countAlivePeers()
	bs := srv.createBroadcastSession(quorum)
	bs.QueueDirectMsg(outMsg.PktType, outMsg)
}
