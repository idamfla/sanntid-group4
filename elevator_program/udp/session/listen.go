package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"time"
)

func (ses *Session) listen(behavior SessionBehavior) {
	defer ses.wg.Done()

	ticker := time.NewTicker(udp.RETRY_INTERVAL)
	defer ticker.Stop()

	retryCounter := 0

	for {
		select {
		case <-ses.stop:
			return

		case pkt, ok := <-ses.packetInCh:
			if !ok {
				// Channel closed, stop the session
				fmt.Printf("Session %d recvCh channel closed, stopping\n", ses.ID)
				return
			}
			retryCounter = 0
			ticker.Reset(udp.RETRY_INTERVAL)

			behavior.HandlePacket(pkt)

		case <-ticker.C:
			if ses.hasLastPkt {
				fmt.Println("No answer so far ... retry:", retryCounter)
				ses.sendRetry(ses.lastOutMsg)
				retryCounter++
				if retryCounter > udp.MAX_RETRIES {
					fmt.Printf("Session %d: receiver seems dead, stopping retryCounter\n", ses.ID)
					ses.queueWhoIsAliveMsg() // TODO test that this actually work as fault tol ...
					ses.requestClose()
					return
				}
			}
		}
	}
}

func (ses *Session) ReceivePacket(pkt packet.Packet) {
	select {
	case ses.packetInCh <- pkt:
	default:
		fmt.Println("Session mailbox is full, could not receive new packet")
	}
}
