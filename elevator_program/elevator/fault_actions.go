package elevator

import (	"fmt"
            "time"
            "elevator_program/elevio"
	        "elevator_program/fault"
            "elevator_program/types"
)

// ------------------------- Utility helpers -------------------------- //

/*
func (e *Elevator) hasActiveCabRequests() bool {
	for _, status := range e.System.Elevators[e.Id].CabRequests {
		if status != types.NotActive {
			return true
		}
	}
	return false
}
*/


func (e *Elevator) shouldRestartAfterOffline() bool {
    if !(e.offline && e.scheduleRestart) {
        return false
    }
/*
    if e.hasActiveCabRequests() {
        return false
    }

    if e.hasActiveHallRequests(){
        return false
        }
*/



    if e.System.Elevators[e.Id].State == types.ES_Moving {
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
    fmt.Printf("Elevator %d: cab queue finished while offline, restarting\n", e.Id)
        e.restartScheduled = false

    go func() {
        time.Sleep(500 * time.Millisecond)
        fault.RestartSelf()
    }()
}


// ------------------------- Fault handlers -------------------------- //

func (e *Elevator) handleMotorStopFault(reason string) {
	fmt.Printf("Motor stop fault in elevator %d: %s\n", e.Id, reason)

	e.stopLocally()
	e.enterOfflineMode()
	go func() {
		time.Sleep(500 * time.Millisecond)
		fault.RestartSelf()
	}()
}

func (e *Elevator) handleMasterSuspected(reason string) {
    fmt.Printf("Elevator %d suspects master failure: %s\n", e.Id, reason)
    e.connectedToMaster = false
	e.runElection(reason)
}


func (e *Elevator) handleNetworkFault(reason string) {
	fmt.Printf("Network fault in elevator %d: %s\n", e.Id, reason)

    e.restartScheduled = true
}

func (e *Elevator) handlePeerDead(peerID string) {
    fmt.Println("Peer dead:", peerID)
    if e.faultTolerance != nil {
		e.faultTolerance.RemovePeer(peerID)
	}

	delete(e.System.Elevators, peerID)

	if peerID == e.currentMasterID || peerID < e.Id || e.IsMaster {
		e.runElection("peer dead")
	}

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
    tempElevator := e.System.Elevators[e.Id]
	tempElevator.State = types.ES_Idle
    e.System.Elevators[e.Id] = tempElevator

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

func (e *Elevator) FT_SeenPeer(peerID string) {
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


