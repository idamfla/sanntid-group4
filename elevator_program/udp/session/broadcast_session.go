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
		bs.startAckTimer()
	case packet.PKT_T_IAmMaster:
		bs.startAckTimer()
	case packet.PKT_T_BroadcastUpdate:
		bs.QueueElevatorStateTask()
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
	peerID := pkt.Header.SenderAddr

	bs.addResponder(peerID)
	isQuorumReached := bs.countResponders() >= bs.expectedResponses

	switch h.PktType {
	case packet.PKT_T_IAmAlive:
		bs.hasLastPkt = false

	case packet.PKT_T_WhoIsMaster:
		fmt.Println("Collecting master responses ...")

		bs.SendReply(packet.PKT_T_IAmAlive)

		bs.mu.Lock()
		if !bs.electionStarted {
			bs.electionStarted = true
			bs.wg.Add(1)
			go bs.electMaster()
		}
		bs.mu.Unlock()

	case packet.PKT_T_IAmMaster:
		select {
		case bs.masterFound <- struct{}{}:
		default:
		}

		bs.stopAckTimer()

		bs.SendReply(packet.PKT_T_MasterAck)
		bs.hasLastPkt = false
		bs.scheduleSessionClose()

	case packet.PKT_T_MasterAck:
		if bs.tx.IsMaster() {
			fmt.Printf("MstrAck: %d/%d\n", bs.countResponders(), bs.expectedResponses)
			if isQuorumReached {
				bs.seq++
				bs.hasLastPkt = false
				bs.stopAckTimer()
				bs.requestClose()
			}
		}

	case packet.PKT_T_BroadcastAck:
		fmt.Printf("bcAck: %d/%d\n", bs.countResponders(), bs.expectedResponses)
		if isQuorumReached {
			bs.seq++
			bs.stopAckTimer()

			// TODO elevator should receive and then start a broadcast session where it send the packet to everyone
			bs.notifyTaskReady()

			bs.SendReply(packet.PKT_T_BroadcastCommit)
			bs.resetResponders()
		}
	case packet.PKT_T_BroadcastDone:
		fmt.Printf("bcDone: %d/%d\n", bs.countResponders(), bs.expectedResponses)
		if isQuorumReached {
			bs.seq++
			bs.hasLastPkt = false
			bs.stopRemoteCommitTimer()
			bs.requestClose()
		}
	}

	return nil
}

func (bs *BroadcastSession) startAckTimer() {
	bs.broadcastAckTimer.Restart(udp.BROADCAST_ACK_TIMEOUT, func() {
		fmt.Println("Not enough elevators received the data in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
		bs.requestClose()
	})
}

func (bs *BroadcastSession) stopAckTimer() {
	bs.broadcastAckTimer.Stop()
}

func (bs *BroadcastSession) startRemoteCommitTimer() {
	bs.broadcastCommitTimer.Restart(udp.BROADCAST_COMMIT_TIMEOUT, func() {
		fmt.Println("Not enough elevators completed the task in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
		bs.requestClose()
	})
}

func (bs *BroadcastSession) stopRemoteCommitTimer() {
	bs.broadcastCommitTimer.Stop()
}

func (bs *BroadcastSession) addResponder(addr string) bool {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	if !bs.responders[addr] {
		bs.responders[addr] = true
		return true
	}
	return false
}

func (bs *BroadcastSession) resetResponders() {
	bs.mu.Lock()
	bs.responders = make(map[string]bool)
	bs.responders[bs.selfAddr] = true
	bs.mu.Unlock()
}

func (bs *BroadcastSession) countResponders() int {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return len(bs.responders)
}

func (bs *BroadcastSession) electMaster() {
	defer bs.wg.Done()

	timer := time.NewTimer(udp.MASTER_ELECTION_TIMEOUT)
	defer timer.Stop()

	select {
	case <-bs.masterFound:
		fmt.Println("Master already exists, stopping election")
		return

	case <-timer.C:
		fmt.Println("No master found, electing...")

		bs.mu.Lock()

		if len(bs.responders) == 0 {
			bs.mu.Unlock()
			return
		}

		lowest := ""
		for addr := range bs.responders {
			if lowest == "" || addr < lowest {
				lowest = addr
			}
		}

		isMaster := lowest == bs.selfAddr

		bs.mu.Unlock()

		fmt.Println("Lowest:", lowest)

		if isMaster {
			fmt.Println(bs.selfAddr, "is the new master")
			bs.SendReply(packet.PKT_T_IAmMaster)
			bs.expectedResponses = bs.countResponders() - 1
		}

	case <-bs.stop:
		return
	}
}
