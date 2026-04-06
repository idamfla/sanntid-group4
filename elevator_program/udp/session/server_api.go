package session

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"net"
)

type ServerAPI interface {
	Send(ses *Session, pktType packet.PacketType, outMsg packet.OutgoingMessage) error
	IsMaster() bool
	QueueWhoIsAliveMsg()
	QueueIAmMasterMsg()
	QueueSyncCompleteMsg(outPkt packet.OutgoingMessage)
	QueueElectedMasterMsg(masterAddr string)
	QueueElevatorTask(eMsg message.ElevatorMessage, elevDone chan<- struct{}, taskReady <-chan struct{})
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

func (ses *Session) isMaster() bool      { return ses.srv.IsMaster() }
func (ses *Session) queueWhoIsAliveMsg() { ses.srv.QueueWhoIsAliveMsg() }
func (ses *Session) queueIamMasterMsg()  { ses.srv.QueueIAmMasterMsg() }

func (ses *Session) queueSyncCompleteMsg(outPkt packet.OutgoingMessage) {
	ses.srv.QueueSyncCompleteMsg(outPkt)
}

// expects a response/completion from elevator
func (ses *Session) queueElevatorTask(eMsg message.ElevatorMessage) {
	ses.srv.QueueElevatorTask(eMsg, ses.elevDone, ses.taskReady)
}
