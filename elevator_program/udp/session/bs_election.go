package session

import (
	"elevator_program/message"
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"sync"
	"time"
)

// Election handles the master election logic
type Election struct {
	started     bool
	mu          sync.Mutex
	masterFound chan struct{}
}

// Start launches the election if not already started
func (e *Election) Start(ws *WhoIsMasterBroadcast) {
	e.mu.Lock()
	if e.started {
		e.mu.Unlock()
		return
	}
	e.started = true
	e.mu.Unlock()

	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		e.run(ws)
	}()
}

// run contains the election logic
func (e *Election) run(ws *WhoIsMasterBroadcast) {
	timer := time.NewTimer(udp.MASTER_ELECTION_TIMEOUT)
	defer timer.Stop()

	select {
	case <-ws.stop:
		return

	case <-e.masterFound:
		fmt.Println("Master already exists, stopping election")
		return

	case <-timer.C:
		e.runElection(ws)
	}
}

func (e *Election) runElection(ws *WhoIsMasterBroadcast) {
	fmt.Printf("No master found, electing ... There are %d candidates\n", ws.countTotalResponders())

	lowest, err := ws.runElection()
	if err != nil {
		fmt.Println(err)
		return
	}

	amMaster := lowest == ws.selfAddr

	fmt.Printf("New master elected: %s\n", lowest)

	if amMaster {
		// I am the new master
		ws.expectedResponses = ws.countResponders()
		ws.resetResponders()
		ws.queueReply(packet.PKT_T_IAmMaster)
	} else {
		select {
		case <-e.masterFound:
			fmt.Println("Master found in another election, aborting")
			return
		default:
			ws.QueueDirectMsg(packet.PKT_T_ElectedMasterIs, message.ElevatorMessage{ID: lowest, Addr: lowest})
		}
	}
}
