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
func (srv *Server) handleOutPkt(outMsg packet.OutgoingMessage) {
	defer srv.wg.Done()
	switch outMsg.PktType {
	case packet.PKT_T_WhoIsAlive, packet.PKT_T_IAmMaster:
		// srv.dispatchWhoIsMaster()
		srv.dispatchMasterElectionMsg(outMsg)

	case packet.PKT_T_BroadcastUpdate, packet.PKT_T_SyncMsg:
		// case packet.PKT_T_BroadcastUpdate:
		srv.dispatchBroadcastUpdate(outMsg)

	case packet.PKT_T_RequestTaskExecution:
		srv.dispatchToMasterMsg(outMsg)
	}
}

func (srv *Server) dispatchMasterElectionMsg(outMsg packet.OutgoingMessage) {
	ws := srv.getOrCreateMasterElectionSession()
	ws.QueueDirectMsg(outMsg.PktType, outMsg)
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

func (srv *Server) dispatchBroadcastUpdate(outMsg packet.OutgoingMessage) {
	if !srv.IsMaster() {
		fmt.Println(srv.GetAlias(), "is not master, can't broadcast like one ...")
	}

	srv.startStateBroadcast(outMsg)
}

func (srv *Server) startSession(remoteAddr *net.UDPAddr, outMsg packet.OutgoingMessage) { // TODO move some parts into createSession, rest is a queueMsg or something
	ses := srv.createSession(remoteAddr, nil)
	ses.QueueDirectMsg(outMsg.PktType, outMsg)
}

// Initiate the broadcast message chain
func (srv *Server) startStateBroadcast(outMsg packet.OutgoingMessage) { // TODO could probably just take a outPkt and then extract the pktType and eMsg
	quorum := srv.countActivePeers()
	bs := srv.createBroadcastSession(quorum)
	bs.QueueDirectMsg(outMsg.PktType, outMsg)
}

func (srv *Server) QueueWhoIsAliveMsg() {
	srv.QueueMessage(nil, packet.PROTO_PKT_T_WhoIsAlive, message.ElevatorMessage{})
}

func (srv *Server) QueueElectedMasterMsg(masterAddr string) {
	peer, exists := srv.getPeer(masterAddr)
	if !exists {
		fmt.Println("Elected master does not exist ...")
		srv.QueueWhoIsAliveMsg()
		return
	}

	id, addr, _, _, _, _ := peer.Snapshot()

	srv.QueueMessage(nil, packet.PROTO_PKT_T_ElectedMasterIs, message.ElevatorMessage{ID: id, Addr: addr.String()})
}

func (srv *Server) QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, eMsg message.ElevatorMessage) {
	pktType := packet.PacketType(protoPktType)

	select {
	case srv.outgoingMsgCh <- packet.OutgoingMessage{
		Origin: packet.Identity{
			Identifier: srv.GetRecvString(),
			Alias:      srv.GetAlias(),
		},
		RemoteAddr: remoteAddr,
		PktType:    pktType,
		EMsg:       eMsg,
	}:
	default:
		fmt.Println("Can't queue message, servers messageQueue is full")
	}
}

// --- helper ---
func (srv *Server) isLocalAddr(addr *net.UDPAddr) bool {
	local := srv.getRecvUDPAddr()
	return addr.IP.Equal(local.IP) && addr.Port == local.Port
}
