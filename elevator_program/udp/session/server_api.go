package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"net"
)

type ServerAPI interface {
	Send(ses *Session, pktType packet.PacketType, outMsg packet.OutgoingMessage) error
	QueueWhoIsAliveMsg()
	QueueIAmMasterMsg()
	QueueElectedMasterMsg(masterAddr string)
	QueueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}, taskReady <-chan struct{})
	IsMaster() bool
	GetRecvString() string
	GetBroadcastAddr() *net.UDPAddr
	GetCloseReqCh() chan uint32
}

func (ses *Session) send(outMsg packet.OutgoingMessage) error {
	ses.seq++
	ses.lastOutMsg = outMsg
	return ses.srv.Send(
		ses,
		outMsg.PktType,
		outMsg,
	)
}

func (ses *Session) sendRetry(outMsg packet.OutgoingMessage) error {
	return ses.srv.Send(
		ses,
		outMsg.PktType,
		outMsg)
}

func (ses *Session) queueWhoIsAliveMsg() { ses.srv.QueueWhoIsAliveMsg() }
func (ses *Session) queueIamMasterMsg()  { ses.srv.QueueIAmMasterMsg() }
func (ses *Session) isMaster() bool      { return ses.srv.IsMaster() }
func (ses *Session) queueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}) {
	ses.srv.QueueElevatorTask(eMsg, elevDone, ses.taskReady)
}
