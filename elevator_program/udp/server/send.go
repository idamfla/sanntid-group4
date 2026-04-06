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
	// eMsg message.ElevatorMessage,
	outMsg packet.OutgoingMessage,
) error {
	senderAddr := srv.GetRecvString()

	data, err := ses.GenerateDataPacket(senderAddr, pktType, outMsg) // TODO needs to include the origin
	// data, err := ses.GenerateDataPacket(senderAddr, pktType, eMsg) // TODO needs to include the origin
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
	defer srv.wg.Done()
	switch outMsg.PktType {
	case packet.PKT_T_WhoIsAlive, packet.PKT_T_IAmMaster:
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

	if outMsg.PktType == packet.PKT_T_IAmMaster {
		srv.setSelfAsMaster()
		srv.setSynced()
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
	ses := srv.getOrCreateSession(nil, &outMsg.Origin.ID)
	srv.updateSyncFromMsg(outMsg)
	ses.QueueDirectMsg(outMsg.PktType, outMsg)
}

func (srv *Server) dispatchToMasterMsg(outMsg packet.OutgoingMessage) {
	mstr := srv.getMasterPeer()
	if mstr == nil {
		fmt.Println(srv.GetAlias(), "dosen't know who master is") // TODO remove later,
		srv.QueueWhoIsAliveMsg()
		return
	}

	srv.startSession(mstr.Addr, outMsg)
}

// --- start sessions ---

func (srv *Server) startSession(remoteAddr *net.UDPAddr, outMsg packet.OutgoingMessage) { // TODO move some parts into createSession, rest is a queueMsg or something
	ses := srv.createSession(remoteAddr, nil)
	ses.QueueDirectMsg(outMsg.PktType, outMsg)
}

// Initiate the broadcast message chain
func (srv *Server) startStateBroadcast(outMsg packet.OutgoingMessage) { // TODO could probably just take a outMsg and then extract the pktType and eMsg
	quorum := srv.countAlivePeers()
	bs := srv.createBroadcastSession(quorum)
	bs.QueueDirectMsg(outMsg.PktType, outMsg)
}
