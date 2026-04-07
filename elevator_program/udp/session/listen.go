package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"elevator_program/utilities"
	"fmt"
	"time"
)

func (ses *Session) listen(behavior SessionBehavior) {
	defer ses.WgDone()

	ticker := time.NewTicker(udp.RETRY_INTERVAL)
	defer ticker.Stop()

	retryCounter := 0

	for {
		select {
		case <-ses.stopCh():
			return

		case pkt, ok := <-ses.packetInCh:
			if !ok {
				// Channel closed, stop the session
				fmt.Printf("Session %d recvCh channel closed, stopping\n", ses.ID)
				return
			}
			retryCounter = 0
			utilities.ResetTicker(ticker, udp.RETRY_INTERVAL)

			behavior.HandleIncPkt(pkt)

		case <-ticker.C:
			rCounter, ok := ses.handleRetry(retryCounter)
			if !ok {
				return
			}

			retryCounter = rCounter

		}
	}
}

func (ses *Session) handleRetry(retryCounter int) (int, bool) {
	if ses.hasLastMsg() {
		fmt.Println("No answer so far ... retry:", retryCounter)
		ses.sendRetry(ses.getLastOutMsg())

		retryCounter++

		if retryCounter > udp.MAX_RETRIES {
			fmt.Printf("Session %d: receiver seems dead, stopping retryCounter\n", ses.ID)
			ses.queueWhoIsAliveMsg() // TODO test that this actually work as fault tol ...
			ses.requestClose()
			return retryCounter, false
		}
	}

	return retryCounter, true
}

func (ses *Session) ReceivePacket(pkt packet.Packet) {
	select {
	case ses.packetInCh <- pkt:
	default:
		fmt.Println("Session mailbox is full, could not receive new packet")
	}
}
