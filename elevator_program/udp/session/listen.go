package session

import (
	"elevator_program/udp"
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
				ses.sendRetry(ses.lastOutPkt)
				retryCounter++
				if retryCounter > udp.MAX_RETRIES {
					fmt.Printf("Session %d: receiver seems dead, stopping retryCounter\n", ses.ID)
					ses.QueueWhoIsAliveMsg() // TODO test that this actually work as fault tol ...
					ses.requestClose()
					return
				}
			}
		}
	}
}
