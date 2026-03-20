package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

type StateBroadcast struct {
	*BaseBroadcastSession
}

func NewStateBroadcast(
	id uint32,
	selfAddr string,
	addr *net.UDPAddr,
	closeReq chan<- uint32,
	tx PacketSender,
	expected int,
) *StateBroadcast {
	sbs := &StateBroadcast{
		BaseBroadcastSession: NewBaseBroadcastSession(id, selfAddr, addr, closeReq, tx, expected),
	}

	return sbs
}

func (sbs *StateBroadcast) Start() {
	sbs.wg.Add(2)
	go sbs.listen(sbs)
	go sbs.sendLoop(sbs)
	fmt.Printf("State broadcast session %d started\n", sbs.ID)
}

func (sbs *StateBroadcast) Close() {
	sbs.BaseBroadcastSession.Close()

}

func (sbs *StateBroadcast) SendReply(pkt packet.PacketType) { sbs.Session.SendReply(pkt) }

func (sbs *StateBroadcast) ReceivePacket(pkt packet.Packet) { sbs.Session.ReceivePacket(pkt) }

func (sbs *StateBroadcast) QueueStateBSUpdateMsg(pktType packet.PacketType, eMsg message.ElevatorMessage) {
	var pktT packet.PacketType
	switch pktType {
	case packet.PKT_T_SyncComplete:
		pktT = packet.PKT_T_SyncComplete
	default:
		pktT = packet.PKT_T_BroadcastUpdate
	}

	sbs.QueueDirectMsg(pktT, eMsg)
}

func (sbs *StateBroadcast) QueueWhoIsMasterMsg() {
	sbs.Session.QueueWhoIsMasterMsg()
}

func (sbs *StateBroadcast) OnSend(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_BroadcastUpdate, packet.PKT_T_SyncComplete:
		sbs.QueueElevatorStateTask()
		sbs.startResponseTimer()
	case packet.PKT_T_BroadcastCommit, packet.PKT_T_SyncCommit:
		sbs.startResponseTimer()
	}

}

func (sbs *StateBroadcast) HandlePacket(pkt packet.Packet) error {
	h := pkt.Header
	pktType := h.PktType
	peerID := h.SenderAddr

	if h.Seq != sbs.seq+1 {
		err := fmt.Errorf("Session %d: seq mismatch (got %d, expected %d), retrying last packet\n",
			sbs.ID, h.Seq, sbs.seq+1)
		fmt.Println(err)
		return err

	}

	sbs.seq = h.Seq
	sbs.addResponder(peerID)

	isQuorum := sbs.countResponders() >= sbs.expectedResponses
	switch pktType {
	case packet.PKT_T_BroadcastAck, packet.PKT_T_SyncAck:
		fmt.Printf("bcAck: %d/%d\n", sbs.countResponders(), sbs.expectedResponses)

		if isQuorum {
			sbs.handleStateBSAck(pktType)
		}

	case packet.PKT_T_BroadcastDone, packet.PKT_T_SyncDone:
		fmt.Printf("bcDone: %d/%d\n", sbs.countResponders(), sbs.expectedResponses)

		if isQuorum {
			sbs.handleStateBSDone()
		}
	}
	return nil
}

func (sbs *StateBroadcast) handleStateBSAck(pktType packet.PacketType) {
	sbs.stopResponseTimer()
	sbs.notifyTaskReady()

	switch pktType {
	case packet.PKT_T_BroadcastAck:
		sbs.SendReply(packet.PKT_T_BroadcastCommit)
	case packet.PKT_T_SyncAck:
		sbs.SendReply(packet.PKT_T_SyncCommit)
	}

	sbs.resetResponders()
}

func (sbs *StateBroadcast) handleStateBSDone() {
	sbs.hasLastPkt = false
	sbs.stopResponseTimer()
	sbs.requestClose()
}
