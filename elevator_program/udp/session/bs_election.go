package session

import (
	"elevator_program/message"
	"elevator_program/udp"
	"fmt"
	"sync"
	"time"
)

type Election struct {
	started     bool
	mu          sync.Mutex
	masterFound chan struct{}
}

// Start launches the election if not already started
func (e *Election) Start(ws *WhoIsAliveBroadcast) {
	e.mu.Lock()
	if e.started {
		e.mu.Unlock()
		return
	}
	e.started = true
	e.mu.Unlock()

	ws.WgAdd(1)
	go e.run(ws)
}

// run contains the election logic
func (e *Election) run(ws *WhoIsAliveBroadcast) {
	defer ws.WgDone()

	defer e.clearStarted()

	timer := time.NewTimer(udp.MASTER_ELECTION_TIMEOUT)
	defer timer.Stop()

	select {
	case <-ws.stopCh():
		return

	case <-e.masterFound:
		fmt.Println("Master already exists, stopping election")
		return

	case <-timer.C:
		e.runElection(ws)
	}
}

func (e *Election) runElection(ws *WhoIsAliveBroadcast) {
	fmt.Printf("No master found, electing ... There are %d candidate(s)\n", ws.countTotalResponders())

	if ws.countResponders() == 0 {
		fmt.Printf(` But no one's listening
 And that's just lonely
`)

		ws.queueElevatorCommand(message.EMSG_T_IAmAlone) // TODO need a type for telling elevator we are offline
		ws.clearMasterPeer()
		ws.queueWhoIsAliveMsg()
		return
	}

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
		ws.queueIamMasterMsg()
	} else {
		select {
		case <-e.masterFound:
			fmt.Println("Master found in another election, aborting")
			return
		default:
			ws.queueElectedMasterMsg(lowest)
		}
	}
}

func (e *Election) clearStarted() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.started = false
}
