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

	if pktType == packet.PKT_T_IAmMaster {
		srv.setSelfAsMaster(true)
		srv.isSynced = true
	}

	return packet.SendPacket(srv.sendConn, remoteAddr, pkt)
}

func (srv *Server) startSession(remoteAddr *net.UDPAddr, pktType packet.PacketType, eMsg message.ElevatorMessage) error {
	if srv.isLocalAddr(remoteAddr) {
		err := fmt.Errorf("Tried to send to oneself %s", remoteAddr.String())
		fmt.Println(err)
		return err
	}

	ses := srv.createSession(remoteAddr, nil)
	ses.QueueDirectMsg(pktType, eMsg)
	// srv.elevatorTaskQueue()
	return nil
}

// Initiate the broadcast message chain
func (srv *Server) startBroadcast(eMsg message.ElevatorMessage) {
	quorum := srv.getPeerCount()
	ses := srv.createBroadcastSession(nil, quorum)

	ses.QueueBroadcastUpdateMsg(eMsg)
}

func (srv *Server) startWhoIsMasterMsg() {
	ses := srv.createBroadcastSession(nil, 0)

	ses.QueueWhoIsMasterMsg()
}

// deciding how to output messages from the server, what type of session should start
func (srv *Server) dispatchMessage(outMsg outgoingMessage) {
	defer srv.wg.Done()
	switch outMsg.PktType {
	case packet.PKT_T_SlaveUpdate, packet.PKT_T_RequestTaskExecution:
		srv.dispatchToMasterMsg(outMsg)
	case packet.PKT_T_BroadcastUpdate:
		srv.dispatchBroadcastUpdate(outMsg)
	case packet.PKT_T_WhoIsMaster:
		srv.dispatchWhoIsMaster()
	}
}

func (srv *Server) dispatchToSlaveMsg(outMsg outgoingMessage) {
	srv.startSession(outMsg.RemoteAddr, outMsg.PktType, outMsg.EMsg)
}

func (srv *Server) dispatchToMasterMsg(outMsg outgoingMessage) {
	mstr := srv.GetMasterPeer()
	if mstr == nil {
		fmt.Println(srv.ID, "dosen't know who master is") // TODO remove later,
		// srv.QueueMessage(nil, packet.PROTO_PKT_T_WhoIsMaster, message.ElevatorMessage{}) // TODO fault tol, FAULT_T_LostMaster, queue who is master
		return
	}
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

	srv.startBroadcast(outMsg.EMsg)
}

func (srv *Server) dispatchWhoIsMaster() {
	srv.mu.Lock()
	if srv.searchingForMaster {
		srv.mu.Unlock()
		return
	}
	srv.searchingForMaster = true
	srv.mu.Unlock()

	if peer := srv.GetMasterPeer(); peer != nil {
		peer.SetMaster(false)
	}
	srv.startWhoIsMasterMsg()
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
