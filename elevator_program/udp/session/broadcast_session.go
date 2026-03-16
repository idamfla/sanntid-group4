package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/udp/timer"
	"fmt"
	"net"
	"sync"
	"time"
)

type BroadcastSession struct {
	*Session // embed base Session

	responsesReceived    int // how many elevators replied
	expectedResponses    int // how many must reply before we commit
	broadcastAckTimer    *timer.Timer
	broadcastCommitTimer *timer.Timer
	mu                   sync.Mutex // protect counters
}

func NewBroadcastSession(
	id uint32,
	addr *net.UDPAddr,
	closeReq chan<- uint32,
	elev chan<- ElevatorPacket,
	tx PacketSender,
	expected int,
) *BroadcastSession {
	bs := &BroadcastSession{
		Session:              NewSession(id, addr, closeReq, elev, tx),
		expectedResponses:    expected,
		broadcastAckTimer:    timer.NewTimer(),
		broadcastCommitTimer: timer.NewTimer(),
	}
	return bs
}

func (bs *BroadcastSession) Start() {
	bs.wg.Add(2)
	go bs.listen(bs)
	go bs.sendLoop(bs)
	fmt.Printf("Broadcast session %d started\n", bs.ID)
}

func (bs *BroadcastSession) Close() {
	bs.closeOnce.Do(func() {
		// Stop normal session timers
		bs.Session.remoteCommitTimer.Stop()
		bs.Session.shutdownDelayTimer.Stop()

		// Stop broadcast-specific timers
		if bs.broadcastAckTimer != nil {
			bs.broadcastAckTimer.Stop()
		}
		if bs.broadcastCommitTimer != nil {
			bs.broadcastCommitTimer.Stop()
		}

		// Stop base session goroutines
		close(bs.Session.stop)
		bs.Session.wg.Wait()

		// Close channels
		close(bs.Session.packetInCh)
		close(bs.Session.outgoingMsgCh)
		close(bs.Session.closeReq)

		// Clear pending packet
		bs.Session.pendingPkt = nil

		fmt.Printf("BroadcastSession %d closed\n", bs.ID)
	})
}

func (bs *BroadcastSession) OnSend(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_BroadcastUpdate:
		bs.startAckTimer()
	case packet.PKT_T_BroadcastCommit:
		bs.startRemoteCommitTimer()
	}

}

func (bs *BroadcastSession) ReceivePacket(pkt packet.Packet) {
	bs.packetInCh <- pkt
}

func (bs *BroadcastSession) HandlePacket(pkt packet.Packet) error {
	h := pkt.Header

	bs.mu.Lock()
	bs.responsesReceived++
	quorumReached := bs.responsesReceived >= bs.expectedResponses // TODO is it okay that the quorum is not the same amount as active elevators??
	bs.mu.Unlock()

	switch h.PktType {
	case packet.PKT_T_BroadcastUpdateAck:
		fmt.Printf("bcAck: %d/%d\n", bs.responsesReceived, bs.expectedResponses)
		if quorumReached {
			bs.seq++
			bs.stopAckTimer()
			bs.sendToElevator(bs.pendingPkt)
			bs.sendReply(packet.PKT_T_BroadcastCommit)
			bs.responsesReceived = 0
		}
	case packet.PKT_T_BroadcastDone:
		fmt.Printf("bcDone: %d/%d\n", bs.responsesReceived, bs.expectedResponses)
		if quorumReached {
			bs.seq++
			bs.pendingPkt = nil
			bs.stopRemoteCommitTimer()
			bs.requestClose()
		}
	}

	return nil
}

func (bs *BroadcastSession) startAckTimer() {
	bs.broadcastAckTimer.Restart(udp.BROADCAST_ACK_TIMEOUT*time.Second, func() {
		fmt.Println("Not enought elevators received the data in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
	})
}

func (bs *BroadcastSession) stopAckTimer() {
	bs.broadcastAckTimer.Stop()
}

func (bs *BroadcastSession) startRemoteCommitTimer() {
	bs.broadcastCommitTimer.Restart(udp.BROADCAST_COMMIT_TIMEOUT*time.Second, func() {
		fmt.Println("Not enought elevators completed the task in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
	})
}

func (bs *BroadcastSession) stopRemoteCommitTimer() {
	bs.broadcastCommitTimer.Stop()
}
