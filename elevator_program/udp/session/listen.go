package session

import (
	"elevator_program/udp"
	"fmt"
	"time"
)

func (ses *Session) listen(behavior SessionBehavior) {
	defer ses.wg.Done()

	ticker := time.NewTicker(udp.RETRY_FREQUENCY * time.Second)
	defer ticker.Stop()
	// lastSeen := ticker
	retransmissions := 0

	for {
		select {
		case incPkt, ok := <-ses.recvCh:
			if !ok {
				// Channel closed, stop the session
				fmt.Printf("Session %d recvCh channel closed, stopping\n", ses.ID)
				return
			}
			retransmissions = 0
			behavior.HandlePacket(incPkt)
			// ses.handlePacket(incPkt)
		case <-ticker.C:
			// ses.retransmitt()
			retransmissions++
			if retransmissions > udp.MAX_RETRIES {
				fmt.Printf("Session %d: receiver seems dead, stopping retransmissions\n", ses.ID)
				return
			}
		case <-ses.stop:
			return

		}
	}
}
