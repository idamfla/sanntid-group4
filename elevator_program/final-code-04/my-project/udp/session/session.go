package session

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/udp/peerinfo"
	"elevator_program/udp/timer"
	"fmt"
	"net"
	"sync"
)

const (
	CHANNEL_BUF = 32
)

type PacketSender interface {
	Send(remoteAddr *net.UDPAddr, seq uint32, sessionID uint32, msgType packet.PacketType, eMsg message.ElevatorMessage) error
	QueueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}, taskReady <-chan struct{})
	QueueMessage(remoteAddr *net.UDPAddr, protoPktType packet.ProtocolPacketType, eMsg message.ElevatorMessage)
	IsMaster() bool
	GetMasterPeer() *peerinfo.PeerInfo
}

type SessionBehavior interface {
	HandlePacket(pkt packet.Packet) error
	OnSend(pktType packet.PacketType)
}

type Session struct {
	ID       uint32
	peerAddr *net.UDPAddr
	peerID   string

	seq uint32

	pendingPkt *packet.Packet
	lastOutPkt outgoingMessage
	hasLastPkt bool

	packetInCh    chan packet.Packet
	outgoingMsgCh chan outgoingMessage

	responseTimer *timer.Timer

	elevDone  chan struct{}
	taskReady chan struct{}
	tx        PacketSender

	closeReq  chan<- uint32
	stop      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once
}

func NewSession(id uint32,
	peerAddr *net.UDPAddr,
	closeReq chan<- uint32,
	transmitter PacketSender,
) *Session {
	ses := &Session{
		ID:            id,
		peerAddr:      peerAddr,
		peerID:        peerAddr.String(),
		pendingPkt:    &packet.Packet{},
		lastOutPkt:    outgoingMessage{},
		hasLastPkt:    false,
		packetInCh:    make(chan packet.Packet, CHANNEL_BUF),
		outgoingMsgCh: make(chan outgoingMessage, CHANNEL_BUF),

		responseTimer: timer.NewTimer(),

		elevDone:  make(chan struct{}, 1),
		taskReady: make(chan struct{}, 1),
		tx:        transmitter,

		stop:     make(chan struct{}, CHANNEL_BUF),
		closeReq: closeReq,
	}

	return ses
}

func (ses *Session) Start() {
	ses.wg.Add(2)
	go ses.listen(ses)
	go ses.sendLoop(ses)
}

func (ses *Session) Close() {
	ses.closeOnce.Do(func() {
		ses.stopResponseTimer()

		close(ses.stop)
		ses.wg.Wait()

		close(ses.elevDone)
		close(ses.taskReady)

		ses.pendingPkt = nil
	})
}

func (ses *Session) startResponseTimer() {
	ses.responseTimer.Restart(udp.BROADCAST_ACK_TIMEOUT, func() {
		fmt.Println("Elevator(s) did not respond in time ...")
		ses.QueueWhoIsMasterMsg()
		ses.stopResponseTimer()
		ses.requestClose()
	})
}

func (ses *Session) stopResponseTimer() {
	ses.responseTimer.Stop()
}
