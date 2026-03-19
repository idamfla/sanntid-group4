package elevator

import (
	"elevator_program/elevio"
	"elevator_program/fault"
	"elevator_program/types"
	"fmt"
	"time"
)

// ------------------------- Configs -------------------------- //

type RecoveryConfig struct {
	SoftRestartTimeout     time.Duration
	RecoveryProofTimeout   time.Duration
	MaxSoftRestartAttempts int
}

var DefaultRecoveryConfig = RecoveryConfig{
	SoftRestartTimeout:     1500 * time.Millisecond,
	RecoveryProofTimeout:   3 * time.Second,
	MaxSoftRestartAttempts: 1,
}

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
	if e.IsOnline && !e.scheduleRestart {
		return false
	}

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
	fmt.Printf("Elevator %s: orders finished while offline, attempting recovery\n", e.Id)
	e.scheduleRestart = false
	e.attemptRecovery()

}

// ------------------------- Fault handlers -------------------------- //

func (e *Elevator) handleMotorStopFault(reason string) {
	fmt.Printf("Motor stop fault in elevator %s: %s\n", e.Id, reason)

	e.stopLocally()
	e.enterOfflineMode()
	e.scheduleRestart = true

}

func (e *Elevator) handleMasterSuspected(reason string) {
	fmt.Printf("Elevator %s suspects master failure: %s\n", e.Id, reason)
	e.connectedToMaster = false
}

func (e *Elevator) handleNetworkFault(reason string) {
	fmt.Printf("Network fault in elevator %s: %s\n", e.Id, reason)

	e.enterOfflineMode()
	e.scheduleRestart = true
}

func (e *Elevator) handlePeerDead(peerID string) {
	fmt.Println("Peer dead:", peerID)

	delete(e.System.Elevators, peerID)

	//if peerID == e.currentMasterID || peerID < e.Id || e.IsMaster {
	//e.runElection("peer dead")
	//}

	// TODO senere: reassign hall calls
	// Midlertidig: marker peer dead i elevatorRegistry / fjern den
}

// ------------------------- Mode helpers -------------------------- //
func (e *Elevator) enterOfflineMode() {
	if !e.IsOnline {
		return
	}

	fmt.Println("Entering offline mode (cab-only)")
	e.IsOnline = false

}

func (e *Elevator) exitOfflineMode() {
	if e.IsOnline {
		return
	}

	fmt.Println("Exiting offline mode (back online)")
	e.IsOnline = true
}

func (e *Elevator) stopLocally() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.direction = elevio.MD_Stop
	tempElevator := e.System.Elevators[e.Id]
	tempElevator.State = types.ES_Idle
	e.System.Elevators[e.Id] = tempElevator

}

func (e *Elevator) SoftRestart() {
	fmt.Printf("Elevator %s: soft restart starting\n", e.Id)

	e.stopLocally()
	e.stopRuntimeLoops()

	e.resetRuntimeState(e.numFloors)
	e.RunElevatorProgram()

	fmt.Printf("Elevator %s: soft restart complete\n", e.Id)
}

func (e *Elevator) attemptRecovery() {
	e.recoveryMu.Lock()
	if e.softRestartInProgress {
		e.recoveryMu.Unlock()
		return
	}

	if e.softRestartAttempts >= e.recoveryCfg.MaxSoftRestartAttempts {
		fmt.Printf("Elevator %s: max soft restart attempts reached, hard restarting\n", e.Id)
		e.recoveryMu.Unlock()
		fault.RestartSelf()
		return
	}

	e.softRestartInProgress = true
	e.softRestartAttempts++
	e.recoveryAwaitingProof = true
	e.recoveryVerified = false
	e.lastRecoveryAttempt = time.Now()
	e.recoveryMu.Unlock()

	fmt.Printf("Elevator %s: attempting soft restart\n", e.Id)

	go func() {
		e.SoftRestart()

		time.Sleep(e.recoveryCfg.SoftRestartTimeout)
		deadline := time.Now().Add(e.recoveryCfg.RecoveryProofTimeout)

		for time.Now().Before(deadline) {
			e.recoveryMu.Lock()
			verified := e.recoveryVerified
			e.recoveryMu.Unlock()

			if verified {
				fmt.Printf("Elevator %s: soft restart succeeded\n", e.Id)

				e.recoveryMu.Lock()
				e.softRestartInProgress = false
				e.softRestartAttempts = 0
				e.recoveryAwaitingProof = false
				e.recoveryMu.Unlock()

				return
			}
			time.Sleep(50 * time.Millisecond)
		}

		fmt.Printf("Elevator %s: recovery proof not received, hard restart\n", e.Id)
		fault.RestartSelf()
	}()
}

func (e *Elevator) markRecoveryVerified() {
	e.recoveryMu.Lock()
	defer e.recoveryMu.Unlock()
	if e.recoveryAwaitingProof {
		e.recoveryVerified = true
	}
}
