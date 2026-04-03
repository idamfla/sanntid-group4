package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"net"
)

type ServerAPI interface {
	Send(ses *Session, msgType packet.PacketType, eMsg message.ElevatorMessage) error
	QueueWhoIsAliveMsg()
	QueueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}, taskReady <-chan struct{})
	IsMaster() bool
	SetSelfAsMaster()
	SetIsSynced(isSynced bool)
	GetRecvString() string
	GetBroadcastAddr() *net.UDPAddr
	GetCloseReqCh() chan uint32
}

func (ses *Session) send(outPkt outgoingMessage) error {
	ses.seq++
	ses.lastOutPkt = outPkt
	return ses.srv.Send(
		ses,
		outPkt.PktType,
		outPkt.EMsg,
	)
}

func (ses *Session) sendRetry(outPkt outgoingMessage) error {
	return ses.srv.Send(
		ses,
		outPkt.PktType,
		outPkt.EMsg)
}

func (ses *Session) queueWhoIsAliveMsg() {
	ses.srv.QueueWhoIsAliveMsg()
}

func (ses *Session) queueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}) {
	ses.srv.QueueElevatorTask(eMsg, elevDone, ses.taskReady)
}

func (ses *Session) isMaster() bool {
	return ses.srv.IsMaster()
}

func (ses *Session) setSelfAsMaster() {
	ses.srv.SetSelfAsMaster()
}

func (ses *Session) setIsSynced(isSynced bool) {
	ses.srv.SetIsSynced(isSynced)
}
