package session

import (
	"elevator_program/message"
	"elevator_program/udp"
	"errors"
	"fmt"
	"time"
)

var (
	ErrSessionStopped  = errors.New("SessionStopped")
	ErrElevatorTimeout = errors.New("ElevatorTimeout")
)

// expects a response/completion from elevator
func (ses *Session) queueElevatorRequest() {
	ses.queueElevatorTask(ses.getPendingMsg().EMsg)
}

// fire-and-forget, reponse will appear in another session
func (ses *Session) queueElevatorCommand(eMsgType message.ElevatorMessageType) {
	eMsg := ses.getPendingMsg().EMsg
	eMsg.Addr = ses.peerAddr.String()
	eMsg.EMsgType = eMsgType

	ses.notifyTaskReady()
	ses.srv.QueueElevatorTask(eMsg, nil, ses.taskReady)
}

// Send packet to elevator, block until timeout or elevator complete its task
func (ses *Session) waitForElevatorDone() error {
	timer := time.NewTimer(udp.LOCAL_COMMIT_TIMEOUT)
	defer timer.Stop()

	select { // wait for completion
	case <-ses.stopCh():
		return ErrSessionStopped

	case <-ses.elevDone:
		fmt.Println("Elevator done commiting")
		return nil

	case <-timer.C:
		return ErrElevatorTimeout
	}
}
