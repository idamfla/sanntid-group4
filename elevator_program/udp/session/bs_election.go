package session

import (
	"elevator_program/udp"
	"elevator_program/udp/packet"
	"fmt"
	"sync"
	"time"
)

type Election struct {
	started     bool
	mu          sync.Mutex
	masterFound chan struct{}
}

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
		fmt.Printf("No master found, electing... There are %d candidates\n", ws.countResponders())
		ws.mu.Lock()

		lowest := ""
		for addr := range ws.responders {
			select {
			case <-ws.stop:
				ws.mu.Unlock()
				return
			default:
			}

			if lowest == "" || addr < lowest {
				lowest = addr
			}
		}

		amMaster := lowest == ws.selfAddr
		ws.mu.Unlock()

		if amMaster {
			ws.expectedResponses = ws.countResponders()
			ws.resetResponders()
			ws.SendReply(packet.PKT_T_IAmMaster)
		}

		fmt.Println(ws.selfAddr, "New master elected", lowest, amMaster, ws.expectedResponses)
	}
}
