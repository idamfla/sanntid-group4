package session

import (
	"elevator_program/udp/packet"
	"fmt"
)

type StateBroadcast struct {
	*BaseBroadcastSession
}

func NewStateBroadcast(
	id uint32,
	srv ServerAPI,
	expected int,
) *StateBroadcast {
	sbs := &StateBroadcast{
		BaseBroadcastSession: NewBaseBroadcastSession(id, srv, expected),
	}

	return sbs
}

func (sbs *StateBroadcast) Start() {
	sbs.wg.Add(2)
	go sbs.listen(sbs)
	go sbs.sendLoop(sbs)
}

func (sbs *StateBroadcast) Close() {
	sbs.BaseBroadcastSession.Close()

}

func (sbs *StateBroadcast) GetID() uint32 { return sbs.BaseBroadcastSession.GetID() }

func (sbs *StateBroadcast) ReceivePacket(pkt packet.Packet) {
	sbs.BaseBroadcastSession.ReceivePacket(pkt)
}

func (sbs *StateBroadcast) QueueDirectMsg(pktType packet.PacketType, outMsg packet.OutgoingMessage) { // TODO this should not exsist outside of session ...
	sbs.BaseBroadcastSession.QueueDirectMsg(pktType, outMsg)
}

func (sbs *StateBroadcast) OnSend(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_BroadcastUpdate, packet.PKT_T_SyncComplete:
		sbs.queueElevatorRequest()
		sbs.startResponseTimer()
	case packet.PKT_T_BroadcastCommit, packet.PKT_T_SyncMsgCommit:
		sbs.startResponseTimer()
	}

}

func (sbs *StateBroadcast) HandleIncPkt(pkt packet.Packet) error {
	h := pkt.Header
	pktType := h.PktType
	peerKey := h.SenderAddr

	if h.Seq != sbs.seq+1 {
		err := fmt.Errorf("Session %d: seq mismatch (got %d, expected %d), retrying last packet\n",
			sbs.ID, h.Seq, sbs.seq+1)
		fmt.Println(err)
		return err

	}

	sbs.addResponder(peerKey)

	isQuorum := sbs.countResponders() >= sbs.expectedResponses

	if isQuorum {
		sbs.seq = h.Seq
	}

	fmt.Printf("%s: %d/%d\n", pktType, sbs.countResponders(), sbs.expectedResponses)
	if isQuorum {
		switch pktType {
		case packet.PKT_T_BroadcastAck, packet.PKT_T_SyncMsgAck:
			sbs.handleStateBSAck(pktType)

		case packet.PKT_T_BroadcastDone, packet.PKT_T_SyncComplete:
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
		sbs.queueReply(packet.PKT_T_BroadcastCommit)
	case packet.PKT_T_SyncMsgAck:
		sbs.queueReply(packet.PKT_T_SyncMsgCommit)
	}

	sbs.resetResponders()
}

func (sbs *StateBroadcast) handleStateBSDone() {
	sbs.clearHasLastPacket()
	sbs.stopResponseTimer()
	sbs.requestClose()
}
