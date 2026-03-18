package session

import (
	"elevator_program/udp"
	"fmt"
	"time"
)

func (ses *Session) listen(behavior SessionBehavior) {
	defer ses.wg.Done()

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
		case <-time.After(udp.RETRY_INTERVAL):
			if ses.hasLastPkt {
				ses.sendRetry(ses.lastOutPkt)
				retryCounter++
				if retryCounter > udp.MAX_RETRIES {
					fmt.Printf("Session %d: receiver seems dead, stopping retryCounter\n", ses.ID)
					ses.requestClose() // TODO just have things close, fault tolerence
					return
				}
			}
		case <-ses.stop:
			return
		}
	}
}
