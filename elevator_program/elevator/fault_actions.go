package elevator

import (
	"fmt"
	"slices"
	"time"
)

const motorWatchdogTimeout = 5 * time.Second

func (e *Elevator) resetMotorWatchdog() {
	e.mu.Lock()
	e.motorWatchdogFloor = e.currentFloor
	e.motorWatchdogTimer = time.Now()
	e.mu.Unlock()
}

func (e *Elevator) checkMotorWatchdog() bool {
	e.mu.Lock()
	floor := e.currentFloor
	wdFloor := e.motorWatchdogFloor
	wdTime := e.motorWatchdogTimer
	e.mu.Unlock()

	if floor != wdFloor {
		e.resetMotorWatchdog()
		return true
	}
	return time.Since(wdTime) < motorWatchdogTimeout
}

func (e *Elevator) UpdateWhoIsALive(ipAdresses []string) {
	for ip, elev := range e.System.Elevators {
		if !slices.Contains(ipAdresses, ip) {
			if e.Ip == ip {
				elev.IsAlive = true
			} else {
				fmt.Println("Elevator %d is dead", ip)
				elev.IsAlive = false
			}
		} else {
			elev.IsAlive = true
		}
		e.System.Elevators[ip] = elev
	}
}
