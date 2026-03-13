package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"net"
	"sync"
	"time"
)

type BroadcastSession struct {
	*Session // embed base Session

	responsesReceived   int // how many elevators replied
	expectedResponses   int // how many must reply before we commit
	broacastAckTimer    time.Timer
	broacastCommitTimer time.Timer
	mu                  sync.Mutex // protect counters
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
		Session:           NewSession(id, addr, closeReq, elev, tx),
		expectedResponses: expected,
	}
	return bs
}

func (bs *BroadcastSession) ReceivePacket(incPkt IncomingPacket) {
	bs.recvCh <- incPkt
}

func (bs *BroadcastSession) handlePacket(incPkt IncomingPacket) error {
	pkt := incPkt.Packet
	h := pkt.Header

	bs.mu.Lock()
	bs.responsesReceived++
	quorumReached := bs.responsesReceived >= bs.expectedResponses
	bs.mu.Unlock()

	switch h.PktType {
	case packet.PKT_T_BroadcastAck:
		fmt.Println("BS handle packet triggerd :)!!!")

		if quorumReached {
			bs.sendReply(packet.PKT_T_BroadcastCommit)
			bs.broadcastCommitTimer()
		}
		fmt.Printf("bcAck: %d/%d", bs.responsesReceived, bs.expectedResponses)
	case packet.PKT_T_BroadcastDone:
		if quorumReached {
			bs.requestClose()
		}

	}

	return nil
}

func (bs *BroadcastSession) broadcastAckTimer() {
	bs.remoteCommitTimer.Restart(udp.BROADCAST_ACK_TIMEOUT*time.Second, func() {
		fmt.Println("Not enought elevators completed the task in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
	})
}

func (bs *BroadcastSession) broadcastCommitTimer() {
	bs.remoteCommitTimer.Restart(udp.BROADCAST_COMMIT_TIMEOUT*time.Second, func() {
		fmt.Println("Not enought elevators completed the task in time ...")
		// TODO what now??
		// ses.closeReq <- ses.ID
	})
}

// TODO make broadcast start timeAck and timeCommit when sending broadcast data and broadcast commit respectively
