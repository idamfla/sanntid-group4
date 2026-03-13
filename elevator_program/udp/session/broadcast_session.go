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
	go bs.Listen(bs)
	go bs.SendLoop(bs)
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
		close(bs.Session.recvCh)
		close(bs.Session.sendCh)
		close(bs.Session.closeReq)

		// Clear pending packet
		bs.Session.pendingPkt = nil

		fmt.Printf("BroadcastSession %d closed\n", bs.ID)
	})
}

func (bs *BroadcastSession) OnSend(pktType packet.PacketType) {
	switch pktType {
	case packet.PKT_T_BroadcastData:
		bs.startAckTimer()
	case packet.PKT_T_BroadcastCommit:
		bs.startRemoteCommitTimer()
	}

}

func (bs *BroadcastSession) ReceivePacket(incPkt IncomingPacket) {
	bs.recvCh <- incPkt
}

func (bs *BroadcastSession) HandlePacket(incPkt IncomingPacket) error {
	pkt := incPkt.Packet
	h := pkt.Header

	bs.mu.Lock()
	bs.responsesReceived++
	quorumReached := bs.responsesReceived >= bs.expectedResponses
	bs.mu.Unlock()

	switch h.PktType {
	case packet.PKT_T_BroadcastAck:
		if quorumReached {
			bs.stopAckTimer()
			bs.sendReply(packet.PKT_T_BroadcastCommit)
		}
		fmt.Printf("bcAck: %d/%d", bs.responsesReceived, bs.expectedResponses)
	case packet.PKT_T_BroadcastDone:
		if quorumReached {
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

// TODO seq is off ...
/*
Server A is running...
Server B is running...
New session created: 1516150186
Broadcast session 1516150186 started
A listening on 127.0.0.1:9000
A listening on 0.0.0.0:3000
B listening on 127.0.0.1:9001
B listening on 0.0.0.0:3000
New session created: 1516150186
Session 1516150186 started
B, Session 1516150186:
        sent from : 192.168.50.97:56520
        to        : 192.168.50.255:3000
        reply sock: 127.0.0.1:9000
        pktType   : Broadcast Data
        seq : 1
        pktType : Broadcast Data
        payload : {Content:Hello from A}
A, Session 1516150186:
        sent from : 127.0.0.1:60660
        to        : 127.0.0.1:9000
        reply sock: 127.0.0.1:9001
        pktType   : Broadcast Ack
BS handle packet triggerd :)!!!
bcAck: 1/0B, Session 1516150186:
        sent from : 192.168.50.97:56520
        to        : 192.168.50.255:3000
        reply sock: 127.0.0.1:9000
        pktType   : Broadcast Commit
order of packages is off ... got: 2, expected: 3
        seq : 2
        pktType : Broadcast Commit
        payload : {Content:}
msgCh test: Hello from A
Elevator done commiting
A, Session 1516150186:
        sent from : 127.0.0.1:60660
        to        : 127.0.0.1:9000
        reply sock: 127.0.0.1:9001
        pktType   : Broadcast Done
BroadcastSession 1516150186 closed
Server A removed session: 1516150186
*/
