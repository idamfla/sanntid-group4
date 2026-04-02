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

func (srv *Server) startSession(remoteAddr *net.UDPAddr, pktType packet.PacketType, eMsg message.ElevatorMessage) error {
	if srv.isLocalAddr(remoteAddr) {
		err := fmt.Errorf("Tried to send to oneself %s", remoteAddr.String())
		fmt.Println(err)
		return err
	}

	ses := srv.createSession(remoteAddr, nil)
	ses.QueueDirectMsg(pktType, eMsg)
	return nil
}

// Initiate the broadcast message chain
func (srv *Server) startStateBroadcast(pktType packet.PacketType, eMsg message.ElevatorMessage) {
	quorum := srv.getPeerCount()
	bs := srv.createBroadcastSession(nil, session.BS_T_StateBroadcast, quorum)

	bs.QueueStateBSUpdateMsg(pktType, eMsg)
}

func (srv *Server) startWhoIsMasterMsg() {
	bs := srv.createBroadcastSession(nil, session.BS_T_WhoIsMasterBroadcast, 0)

	bs.QueueWhoIsMasterMsg()
}

// deciding how to output messages from the server, what type of session should start
func (srv *Server) dispatchMessage(outMsg outgoingMessage) {
	defer srv.wg.Done()
	switch outMsg.PktType {
	case packet.PKT_T_WhoIsMaster:
		srv.dispatchWhoIsMaster()

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
	srv.startSession(outMsg.RemoteAddr, outMsg.PktType, outMsg.EMsg)
}

func (srv *Server) dispatchToMasterMsg(outMsg outgoingMessage) {
	srv.mu.Lock()

	mstr := srv.getMasterPeerLocked()
	if mstr == nil {
		fmt.Println(srv.ID, "dosen't know who master is") // TODO remove later,
		// srv.QueueMessage(nil, packet.PROTO_PKT_T_WhoIsMaster, message.ElevatorMessage{}) // TODO fault tol, FAULT_T_LostMaster, queue who is master
		srv.mu.Unlock()
		return
	}
	srv.mu.Unlock()

	srv.startSession(mstr.Addr, outMsg.PktType, outMsg.EMsg)
}

func (srv *Server) dispatchBroadcastUpdate(outMsg outgoingMessage) {
	if !srv.IsMaster() {
		fmt.Println(srv.ID, "is not master, can't broadcast like one ...")
	}

	// if some peers are syncing
	srv.mu.Lock()
	for _, p := range srv.peers {
		if p.Active && !p.IsSynced {
			p.QueueMessage(outMsg.EMsg)
		}
	}
	srv.mu.Unlock()

	srv.startStateBroadcast(packet.PKT_T_BroadcastUpdate, outMsg.EMsg)
}

func (srv *Server) dispatchCatchupDone(outMsg outgoingMessage) {
	if !srv.IsMaster() {
		fmt.Println(srv.ID, "is not master, can't broadcast like one ...")
	}

	srv.startStateBroadcast(packet.PKT_T_SyncComplete, outMsg.EMsg)
}

func (srv *Server) dispatchWhoIsMaster() {
	if srv.isSearchingForMaster() {
		return
	}

	srv.ResetState()
	srv.setMasterSearch()

	srv.mu.Lock()
	if peer := srv.getMasterPeerLocked(); peer != nil {
		peer.SetMaster(false)
	}
	srv.mu.Unlock()

	srv.startWhoIsMasterMsg()
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

func (srv *Server) QueueSyncRequest() {
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
