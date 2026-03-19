package session

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/udp/timer"
	"fmt"
	"net"
)

type StateBroadcast struct {
	*BaseBroadcastSession
	broadcastCommitTimer *timer.Timer
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
		broadcastCommitTimer: timer.NewTimer(),
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
	sbs.closeOnce.Do(func() {
		if sbs.broadcastCommitTimer != nil {
			sbs.broadcastCommitTimer.Stop()
		}

		sbs.BaseBroadcastSession.Close()
	})
}

func (sbs *StateBroadcast) SendReply(pkt packet.PacketType) { sbs.Session.SendReply(pkt) }

func (sbs *StateBroadcast) ReceivePacket(pkt packet.Packet) { sbs.Session.ReceivePacket(pkt) }

func (sbs *StateBroadcast) QueueBroadcastUpdateMsg(eMsg message.ElevatorMessage) {
	sbs.outgoingMsgCh <- outgoingMessage{
		PktType: packet.PKT_T_BroadcastUpdate,
		EMsg:    eMsg,
	}
}

func (sbs *StateBroadcast) QueueWhoIsMasterMsg() {}

func (sbs *StateBroadcast) OnSend(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_BroadcastUpdate:
		sbs.QueueElevatorStateTask()
		sbs.startAckTimer()
	case packet.PKT_T_BroadcastCommit:
		sbs.startRemoteCommitTimer()
	}

}

func (sbs *StateBroadcast) HandlePacket(pkt packet.Packet) error {
	h := pkt.Header
	peerID := pkt.Header.SenderAddr

	if h.Seq != sbs.seq+1 {
		err := fmt.Errorf("Session %d: seq mismatch (got %d, expected %d), retrying last packet\n",
			sbs.ID, h.Seq, sbs.seq+1)
		fmt.Println(err)
		return err

	}

	sbs.seq = pkt.Header.Seq
	sbs.addResponder(peerID)

	isQuorum := sbs.countResponders() >= sbs.expectedResponses
	switch pkt.Header.PktType {
	case packet.PKT_T_BroadcastAck:
		fmt.Printf("bcAck: %d/%d\n", sbs.countResponders(), sbs.expectedResponses)

		if isQuorum {
			sbs.handleBroadcastAck()
		}

	case packet.PKT_T_BroadcastDone:
		fmt.Printf("bcDone: %d/%d\n", sbs.countResponders(), sbs.expectedResponses)

		if isQuorum {
			sbs.handleBroadcastDone()
		}
	}
	return nil
}

func (sbs *StateBroadcast) handleBroadcastAck() {
	sbs.stopAckTimer()
	sbs.notifyTaskReady()
	sbs.SendReply(packet.PKT_T_BroadcastCommit)
	sbs.resetResponders()
}

func (sbs *StateBroadcast) handleBroadcastDone() {
	sbs.hasLastPkt = false
	sbs.stopRemoteCommitTimer()
	sbs.requestClose()
}

func (sbs *StateBroadcast) startRemoteCommitTimer() {
	sbs.broadcastCommitTimer.Restart(udp.BROADCAST_COMMIT_TIMEOUT, func() {
		fmt.Println("Not enough elevators completed the task in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
		sbs.requestClose()
	})
}

func (bs *StateBroadcast) stopRemoteCommitTimer() {
	bs.broadcastCommitTimer.Stop()
}
