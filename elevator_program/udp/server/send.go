package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"elevator_program/udp/session"
	"fmt"
	"net"
)

func (srv *Server) Send(
	ses *session.Session,
	pktType packet.PacketType,
	eMsg message.ElevatorMessage,
) error {
	senderAddr := srv.GetRecvString()

	data, err := ses.GenerateDataPacket(senderAddr, pktType, eMsg)
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
func (srv *Server) handleOutPkt(outMsg outgoingMessage) {
	defer srv.wg.Done()
	switch outMsg.PktType {
	case packet.PKT_T_WhoIsAlive:
		srv.QueueWhoIsAliveMsg()

	case packet.PKT_T_BroadcastUpdate:
		srv.dispatchBroadcastUpdate(outMsg)

	case packet.PKT_T_SlaveUpdate, packet.PKT_T_RequestTaskExecution:
		srv.dispatchToMasterMsg(outMsg)

	case packet.PKT_T_CatchupUpdate, packet.PKT_T_Snapshot:
		srv.dispatchToSlaveMsg(outMsg)

	case packet.PKT_T_SyncComplete:
		srv.dispatchCatchupDone(outMsg)
	}
}

func (srv *Server) dispatchToSlaveMsg(outMsg outgoingMessage) {
	srv.startSession(outMsg.RemoteAddr, outMsg)
}

func (srv *Server) dispatchToMasterMsg(outMsg outgoingMessage) {
	mstr := srv.getMasterPeer()
	if mstr == nil {
		fmt.Println(srv.ID, "dosen't know who master is") // TODO remove later,
		srv.QueueWhoIsAliveMsg()
		return
	}

	srv.startSession(mstr.Addr, outMsg)
}

func (srv *Server) dispatchBroadcastUpdate(outMsg outgoingMessage) {
	if !srv.IsMaster() {
		fmt.Println(srv.ID, "is not master, can't broadcast like one ...")
	}

	// if some peers are syncing // TODO make this universal, no msg when we are trying to sync alive-unsynced-peers
	// srv.mu.Lock()
	// for _, p := range srv.peers {
	// 	if p.Active && !p.IsSynced {
	// 		p.QueueMessage(outMsg.EMsg)
	// 	}
	// }
	// srv.mu.Unlock()

	srv.startStateBroadcast(outMsg)
}

func (srv *Server) dispatchCatchupDone(outMsg outgoingMessage) {
	if !srv.IsMaster() {
		fmt.Println(srv.ID, "is not master, can't broadcast like one ...")
	}

	srv.startStateBroadcast(outMsg)
}

func (srv *Server) startSession(remoteAddr *net.UDPAddr, outMsg outgoingMessage) { // TODO move some parts into createSession, rest is a queueMsg or something
	ses := srv.createSession(remoteAddr, nil)
	ses.QueueDirectMsg(outMsg.PktType, outMsg.EMsg)
}

// Initiate the broadcast message chain
func (srv *Server) startStateBroadcast(outMsg outgoingMessage) { // TODO could probably just take a outPkt and then extract the pktType and eMsg
	quorum := srv.countActivePeers()
	bs := srv.createBroadcastSession(quorum)
	bs.QueueStateBSUpdateMsg(outMsg.PktType, outMsg.EMsg)
}

func (srv *Server) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, eMsg message.ElevatorMessage) {
	pktType := packet.PacketType(protoPktType)

	select {
	case srv.outgoingMsgCh <- outgoingMessage{
		RemoteAddr: remoteAddr,
		PktType:    pktType,
		EMsg:       eMsg,
	}:
	default:
		fmt.Println("Can't queue message, servers messageQueue is full")
	}
}

// TODO syncing flow is changing ...
func (srv *Server) queueSyncRequest() {
	srv.QueueMessage(nil, packet.PROTO_PKT_T_RequestTaskExecution, message.ElevatorMessage{
		ID:       srv.ID,
		Addr:     srv.GetRecvString(),
		EMsgType: message.EMSG_T_NewToChannel,
	})
}

// --- helper ---
func (srv *Server) isLocalAddr(addr *net.UDPAddr) bool {
	local := srv.getRecvUDPAddr()
	return addr.IP.Equal(local.IP) && addr.Port == local.Port
}
