package elevator

import (
	"elevator_program/message"
	"fmt"
)

// nb
// run as go routine
func (e *Elevator) fault_loop() {
	defer e.wg.Done()
	for {
		select {
		case <-e.stop:
			return
		case faultMsg := <-e.FaultMsg:
			e.handleFaultMessage(faultMsg)
		}
	}
}

func (e *Elevator) handleFaultMessage(faultMsg message.FaultMessage) {
	switch faultMsg.FaultType {
	case message.FAULT_T_LostConn:
		e.handleNetworkFault("lost connection")

	case message.FAULT_T_LostMaster:
		e.handleMasterSuspected("lost master")

	case message.FAULT_T_ElevatorFailed:
		e.handleMotorStopFault("elevator failed")

	case message.FAULT_T_BroadcastFailedToRespond:
		if faultMsg.ID == "" {
			fmt.Println("Received FAULT_T_BroadcastFailedToRespond without peer ID")
			return
		}
		e.handlePeerDead(faultMsg.ID)

	case message.FAULT_T_TaskRunningErr:
		e.handleMotorStopFault("task running too long")

	default:
		fmt.Printf("Unknown fault type: %v\n", faultMsg.FaultType)
	}
}
