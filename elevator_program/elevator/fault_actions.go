package elevator

import (
	"elevator_program/elevio"
	"elevator_program/fault"
	"elevator_program/types"
	"fmt"
	"time"
)

// ------------------------- Config ------------------------- //

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

// ------------------------- Types ------------------------- //

type RestartReason string

const (
	RestartReasonNone    RestartReason = ""
	RestartReasonNetwork RestartReason = "network"
	RestartReasonMotor   RestartReason = "motor"
)

// ------------------------- State / utility helpers ------------------------- //

func (e *Elevator) hasActiveCabRequests() bool {
	e.System.Mutex.RLock()
	defer e.System.Mutex.RUnlock()

	elev, ok := e.System.Elevators[e.Id]
	if !ok {
		return false
	}

	for _, status := range elev.CabRequests {
		if status != types.NotActive {
			return true
		}
	}

	return false
}

func (e *Elevator) hasPendingLocalWork() bool {
	e.System.Mutex.RLock()
	defer e.System.Mutex.RUnlock()

	elev, ok := e.System.Elevators[e.Id]
	if !ok {
		return false
	}

	if elev.Target.Floor != -1 {
		return true
	}

	for _, status := range elev.CabRequests {
		if status != types.NotActive {
			return true
		}
	}

	return false
}

func (e *Elevator) shouldMarkRecoveryVerified() bool {
	e.System.Mutex.RLock()
	elev, ok := e.System.Elevators[e.Id]
	e.System.Mutex.RUnlock()

	if !ok {
		return false
	}

	isMoving := elev.State == types.ES_Moving
	isIdle := elev.State == types.ES_Idle
	doorClosed := e.doorState == DS_Closed
	noEmergency := !e.emergencyStop
	noObstruction := !e.obstruction
	noPendingWork := !e.hasPendingLocalWork()

	safeIdle := isIdle && doorClosed && noEmergency && noObstruction && noPendingWork

	return isMoving || safeIdle

}

func (e *Elevator) shouldRestartAfterOffline() bool {

	e.System.Mutex.RLock()
	elev, ok := e.System.Elevators[e.Id]
	e.System.Mutex.RUnlock()

	if !ok {
		return false
	}

	isOffline := e.offline
	restartScheduled := e.scheduleRestart
	isMoving := elev.State == types.ES_Moving
	doorClosed := e.doorState == DS_Closed
	isNetworkRecovery := e.restartReason == RestartReasonNetwork
	hasCabRequests := e.hasActiveCabRequests()

	canRestartNow := isOffline &&
		restartScheduled &&
		!isMoving &&
		doorClosed

	if !canRestartNow {
		return false
	}

	if isNetworkRecovery && hasCabRequests {
		return false
	}

	return true
}

// ------------------------- Fault handlers ------------------------- //

func (e *Elevator) handleMotorStopFault(reason string) {
	fmt.Printf("Motor stop fault in elevator %s: %s\n", e.Id, reason)

	e.restartReason = RestartReasonMotor
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

	e.restartReason = RestartReasonNetwork
	e.enterOfflineMode()
	e.scheduleRestart = true
}

func (e *Elevator) handlePeerDead(peerID string) {
	fmt.Println("Peer dead:", peerID)

	e.System.Mutex.Lock()
	defer e.System.Mutex.Unlock()

	delete(e.System.Elevators, peerID)
}

// ------------------------- Mode helpers ------------------------- //

func (e *Elevator) enterOfflineMode() {
	if e.offline {
		return
	}

	fmt.Println("Entering offline mode (cab-only)")
	e.offline = true
}

func (e *Elevator) setLocalState(state types.ElevatorState) {
	e.System.Mutex.Lock()
	defer e.System.Mutex.Unlock()

	elev, ok := e.System.Elevators[e.Id]
	if !ok {
		return
	}

	elev.State = state
	e.System.Elevators[e.Id] = elev
}

func (e *Elevator) stopLocally() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.direction = elevio.MD_Stop

	e.setLocalState(types.ES_Idle)
}

// ------------------------- Recovery flow helpers ------------------------- //

func (e *Elevator) checkOfflineRestart() {
	if !e.shouldRestartAfterOffline() {
		return
	}

	fmt.Printf("Elevator %s: orders finished while offline, attempting recovery\n", e.Id)
	e.scheduleRestart = false
	e.attemptRecovery()
}

func (e *Elevator) startRecoveryAttempt() {
	e.softRestartInProgress = true
	e.softRestartAttempts++
	e.recoveryAwaitingProof = true
	e.recoveryVerified = false
	e.lastRecoveryAttempt = time.Now()
}

func (e *Elevator) waitForRecoveryProof(deadline time.Time) bool {
	for time.Now().Before(deadline) {
		e.recoveryMu.Lock()
		recoveryVerified := e.recoveryVerified
		e.recoveryMu.Unlock()

		if recoveryVerified {
			return true
		}

		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func (e *Elevator) markRecoveryVerified() {
	e.recoveryMu.Lock()
	defer e.recoveryMu.Unlock()

	if e.recoveryAwaitingProof {
		e.recoveryVerified = true
	}
}

func (e *Elevator) finishSuccessfulRecovery() {
	fmt.Printf("Elevator %s: soft restart succeeded\n", e.Id)

	e.recoveryMu.Lock()
	e.softRestartInProgress = false
	e.softRestartAttempts = 0
	e.recoveryAwaitingProof = false
	e.restartReason = RestartReasonNone
	e.recoveryMu.Unlock()
}

// ------------------------- Recovery actions ------------------------- //

func (e *Elevator) SoftRestart() {
	fmt.Printf("Elevator %s: soft restart starting\n", e.Id)

	e.stopLocally()
	e.stopRuntimeLoops()

	e.resetRuntimeState(e.numFloors)
	e.RunElevatorProgram()

	fmt.Printf("Elevator %s: soft restart complete\n", e.Id)
}

func (e *Elevator) attemptRecovery() {

	if e.restartReason == RestartReasonNetwork {
		fault.RestartSelf()
		return
	}

	e.recoveryMu.Lock()
	recoveryInProgress := e.softRestartInProgress
	maxAttemptsReached := e.softRestartAttempts >= e.recoveryCfg.MaxSoftRestartAttempts

	if recoveryInProgress {
		e.recoveryMu.Unlock()
		return
	}

	if maxAttemptsReached {
		e.recoveryMu.Unlock()
		fmt.Printf("Elevator %s: max soft restart attempts reached, hard restarting\n", e.Id)
		fault.RestartSelf()
		return
	}

	e.startRecoveryAttempt()
	e.recoveryMu.Unlock()

	fmt.Printf("Elevator %s: attempting soft restart\n", e.Id)

	go func() {
		e.SoftRestart()

		deadline := time.Now().Add(e.recoveryCfg.RecoveryProofTimeout)

		recovered := e.waitForRecoveryProof(deadline)

		if recovered {
			e.finishSuccessfulRecovery()
			return
		}

		fmt.Printf("Elevator %s: recovery proof not received, hard restart\n", e.Id)
		fault.RestartSelf()
	}()
}
