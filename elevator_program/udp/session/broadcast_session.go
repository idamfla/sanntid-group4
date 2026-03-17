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
	selfAddr             string
	broadcastAckTimer    *timer.Timer
	broadcastCommitTimer *timer.Timer
	responders           map[string]bool
	masterFound          chan struct{}
	electionStarted      bool
	mu                   sync.Mutex // protect counters
}

func NewBroadcastSession(
	id uint32,
	selfAddr string,
	addr *net.UDPAddr,
	closeReq chan<- uint32,
	tx PacketSender,
	expected int,
) *BroadcastSession {
	bs := &BroadcastSession{
		Session:              NewSession(id, addr, closeReq, tx),
		expectedResponses:    expected,
		selfAddr:             selfAddr,
		broadcastAckTimer:    timer.NewTimer(),
		broadcastCommitTimer: timer.NewTimer(),
		responders:           make(map[string]bool),
		masterFound:          make(chan struct{}, 1),
		electionStarted:      false,
	}

	bs.responders[bs.selfAddr] = true
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
		close(bs.masterFound)

		// Clear pending packet
		bs.Session.pendingPkt = nil

		fmt.Printf("BroadcastSession %d closed\n", bs.ID)
	})
}

func (bs *BroadcastSession) OnSend(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_WhoIsMaster:

	case packet.PKT_T_BroadcastUpdate:
		bs.QueueElevatorTask()
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
	case packet.PKT_T_IAmAlive:
		bs.addResponder(pkt.Header.SenderAddr)

	case packet.PKT_T_WhoIsMaster:
		fmt.Println("Collecting master responses ...")

		bs.addResponder(pkt.Header.SenderAddr)

		bs.SendReply(packet.PKT_T_IAmAlive)

		if !bs.electionStarted {
			bs.electionStarted = true
			bs.wg.Add(1)
			go bs.electMaster()
		}

	case packet.PKT_T_IAmMaster:
		select {
		case bs.masterFound <- struct{}{}:
		default:
		}
		fmt.Println(pkt.Header.SenderAddr, "was elected master")
		bs.scheduleSessionClose()

	case packet.PKT_T_BroadcastAck:
		fmt.Printf("bcAck: %d/%d\n", bs.responsesReceived, bs.expectedResponses)
		if quorumReached {
			bs.seq++
			bs.stopAckTimer()

			// TODO elevator should receive and then start a broadcast session where it send the packet to everyone
			bs.notifyTaskReady()

			bs.SendReply(packet.PKT_T_BroadcastCommit)
			bs.responsesReceived = 0
		}
	case packet.PKT_T_BroadcastDone:
		fmt.Printf("bcDone: %d/%d\n", bs.responsesReceived, bs.expectedResponses)
		if quorumReached {
			bs.seq++
			bs.stopRemoteCommitTimer()
			bs.requestClose()
		}
	}

	return nil
}

func (bs *BroadcastSession) startAckTimer() {
	bs.broadcastAckTimer.Restart(udp.BROADCAST_ACK_TIMEOUT, func() {
		fmt.Println("Not enought elevators received the data in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
	})
}

func (bs *BroadcastSession) stopAckTimer() {
	bs.broadcastAckTimer.Stop()
}

func (bs *BroadcastSession) startRemoteCommitTimer() {
	bs.broadcastCommitTimer.Restart(udp.BROADCAST_COMMIT_TIMEOUT, func() {
		fmt.Println("Not enought elevators completed the task in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
	})
}

func (bs *BroadcastSession) stopRemoteCommitTimer() {
	bs.broadcastCommitTimer.Stop()
}

func (bs *BroadcastSession) addResponder(addr string) {
	bs.mu.Lock()
	bs.responders[addr] = true
	bs.mu.Unlock()
}

func (bs *BroadcastSession) resetResponders() {
	bs.mu.Lock()
	bs.responders = make(map[string]bool)
	bs.mu.Unlock()
}

func (bs *BroadcastSession) electMaster() {
	defer bs.wg.Done()
	select {
	case <-bs.masterFound:
		fmt.Println("Master already exists, stopping election")
		return

	case <-time.After(udp.MASTER_ELECTION_TIMEOUT):
		fmt.Println("No master found, electing...")

		bs.mu.Lock()
		defer bs.mu.Unlock()

		if len(bs.responders) == 0 {
			fmt.Println("No responders")
			return
		}

		// find lowest
		lowest := ""
		for addr := range bs.responders {
			if lowest == "" || addr < lowest {
				lowest = addr
			}
		}

		fmt.Println("Lowest:", lowest)

		if lowest == bs.selfAddr {
			fmt.Println(bs.selfAddr, "is the new master")
			bs.SendReply(packet.PKT_T_IAmMaster)
			bs.setMaster(true)
			bs.scheduleSessionClose()
		}

	case <-bs.stop:
		return
	}
}

func (bs *BroadcastSession) setMaster(isMaster bool) {
	bs.tx.SetMaster(isMaster)
}
