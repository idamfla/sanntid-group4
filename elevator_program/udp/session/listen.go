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
	// lastSeen := ticker
	retryCounter := 0

	for {
		select {
		case pkt, ok := <-ses.packetInCh:
			if !ok {
				// Channel closed, stop the session
				fmt.Printf("Session %d recvCh channel closed, stopping\n", ses.ID)
				return
			}
			retryCounter = 0
			behavior.HandlePacket(pkt)
		case <-ticker.C:
			// ses.retransmitt()
			retryCounter++
			if retryCounter > udp.MAX_RETRIES {
				fmt.Printf("Session %d: receiver seems dead, stopping retryCounter\n", ses.ID)
				return
			}
		case <-ses.stop:
			return
		}
	}
}
