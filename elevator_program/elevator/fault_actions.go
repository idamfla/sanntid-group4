package elevator

import (	"elevator_program/elevio"
            "fmt"
            "time"
)

// ------------------------- Utility helpers -------------------------- //

func (e *Elevator) hasActiveCabRequests() bool {
	for _, status := range e.system.Elevators[e.id].CabRequests {
		if status != NotActive {
			return true
		}
	}
	return false
}

func (e *Elevator) shouldRestartAfterOffline() bool {
    if !e.offline || !e.restartScheduled {
        return false
    }

    if e.hasActiveCabRequests() {
        return false
    }

    if e.elevatorState == ES_Moving {
        return false
    }

    if e.doorState != DS_Closed {
        return false
    }

    return true
}

func (e *Elevator) checkOfflineRestart() {
    if !e.shouldRestartAfterOffline() {
        return
    }
    fmt.Printf("Elevator %d: cab queue finished while offline, restarting\n", e.id)
        e.restartScheduled = false

    go func() {
        time.Sleep(500 * time.Millisecond)
        restartSelf()
    }()
}


// ------------------------- Fault handlers -------------------------- //

func (e *Elevator) handleMotorStopFault(reason string) {
	fmt.Printf("Motor stop fault in elevator %d: %s\n", e.id, reason)

	e.stopLocally()
	e.enterOfflineMode()
	go func() {
		time.Sleep(500 * time.Millisecond)
		restartSelf()
	}()
}

func (e *Elevator) handleMasterSuspected(reason string) {
    fmt.Printf("Elevator %d suspects master failure: %s\n", e.id, reason)
}


func (e *Elevator) handleNetworkFault(reason string) {
	fmt.Printf("Network fault in elevator %d: %s\n", e.id, reason)

    e.restartScheduled = true
}

func (e *Elevator) handlePeerDead(peerID int) {
    fmt.Println("Peer dead:", peerID)
    if !e.isMaster {
        return
    }

    delete(e.system.Elevators, peerID)

    // TODO senere: reassign hall calls
    // Midlertidig: marker peer dead i elevatorRegistry / fjern den
}



// ------------------------- Mode helpers -------------------------- //
func (e *Elevator) enterOfflineMode() {


    if e.offline {
        return
    }


    fmt.Println("Entering offline mode (cab-only)")
    e.offline = true

    for f := 0; f < len(e.system.hallRequests); f++ {
        elevio.SetButtonLamp(elevio.BT_HallUp, f, false)
        elevio.SetButtonLamp(elevio.BT_HallDown, f, false)
    }
}



func (e *Elevator) exitOfflineMode() {

     if !e.offline {
            return
        }


    fmt.Println("Exiting offline mode (back online)")
    e.offline = false
}



func (e *Elevator) stopLocally() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.direction = elevio.MD_Stop
	e.elevatorState = ES_Idle

	if e.faultTolerance != nil {
		e.faultTolerance.SetMotorRunning(false)
	}
}


// -------------------- Fault-manager interface -------------------- //

func (e *Elevator) FT_SeenMaster() {
    if e.faultTolerance != nil {
        e.faultTolerance.SeenMaster()
    }
}

func (e *Elevator) FT_SeenPeer(peerID int) {
    if e.faultTolerance != nil {
        e.faultTolerance.SeenPeer(peerID)
    }
}

func (e *Elevator) FT_SetRoleMaster() {
    if e.faultTolerance != nil {
        e.faultTolerance.SetRoleMaster()
    }
}

func (e *Elevator) FT_SetRoleSlave() {
    if e.faultTolerance != nil {
        e.faultTolerance.SetRoleSlave()
    }
}

func (e *Elevator) BecameMaster() {
    e.isMaster = true
    if e.faultTolerance != nil {
    e.faultTolerance.SetRoleMaster()
    }
}

func (e *Elevator) BecameSlave() {
    e.isMaster = false
    if e.faultTolerance != nil {
        e.faultTolerance.SetRoleSlave()
    }
}
