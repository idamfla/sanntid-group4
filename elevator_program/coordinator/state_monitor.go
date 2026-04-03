package coordinator

import (
	"elevator_program/elevator"
	"fmt"
	"time"
)

func (c *Coordinator) stateMonitor(e *elevator.Elevator) { // TODO I don't understand why but we never get turntoslave.
	// This is also scary they say you are online, but if you are alone you say you are the master but you are still not online
	defer c.wg.Done()
	fmt.Println("STATE MONITOR STARTED")
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	isMaster := false
	// isOnline := false

	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			fmt.Println(isMaster, c.Server.IsMaster())
			if isMaster != c.Server.IsMaster() {
				isMaster = c.Server.IsMaster()
				if isMaster {
					e.TurnToMaster()
					fmt.Println("Jeg er faen meg sjefen \n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
				} else {
					e.TurnToSlave()
					fmt.Println("Jeg er faen meg sjefen, nei søren det er jeg ikke \n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
				}
			}
			// TODO need is online, or maybe not?
		}
	}
}
