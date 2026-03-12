package session

import (
	"elevator_program/udp/packet"
	"net"
	"sync"
)

type BroadcastSession struct {
	*Session // embed base Session

	ResponsesReceived int        // how many elevators replied
	ExpectedResponses int        // how many must reply before we commit
	mu                sync.Mutex // protect counters
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
		ExpectedResponses: expected,
	}
	return bs
}

func (bs *BroadcastSession) handlePacket(incPkt IncomingPacket) error {
	pkt := incPkt.Packet
	h := pkt.Header

	// call base session handler for common logic
	if err := bs.Session.handlePacket(incPkt); err != nil {
		return err
	}

	// broadcast-specific logic
	if h.PktType == packet.PKT_T_BroadcastAck {
		bs.mu.Lock()
		bs.ResponsesReceived++
		// if threshold reached, trigger broadcast commit
		if bs.ResponsesReceived >= bs.ExpectedResponses {
			bs.mu.Unlock()
			bs.sendReply(packet.PKT_T_BroadcastCommit)
			bs.scheduleSessionClose()
			return nil
		}
		bs.mu.Unlock()
	}

	return nil
}

// TODO have broadcast wait until sufficient amount of acks
// make a cooresponding check for broadcast_done_msg
// have a broadcast timeout that is a bit longer ... maybe give a bit more time since more elevators are included in the communication
