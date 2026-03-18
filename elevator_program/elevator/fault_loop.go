package elevator

import (
	"fmt"
	"elevator_program/fault"
)


//nb
// run as go routine
func (e* Elevator) fault_loop() {
    defer e.wg.Done()
    for {
        select{
        case <-e.stop:
            return
        case faultMsg := <- e.faultMsg:
            e.handleFaultMessage(faultMsg)
        }
    }
}

func (e*Elevator) handleFaultMessage(faultMsg FaultMessage) {
    switch faultMsg.FaultType {
    case fault.FAULT_T_LostConn:
        e.handleNetworkFault("lost connection")

    case fault.FAULT_T_LostMaster:
        e.handleMasterSuspected("lost master")

    case fault.FAULT_T_ElevatorFailed:
        e.handleMotorStopFault("elevator failed")

    case fault.FAULT_T_BroadcastFailedToRespond:
        if faultMsg.ID == "" {
            fmt.Println("Received FAULT_T_BroadcastFailedToRespond without peer ID")
            return
        }
        e.handlePeerDead(faultMsg.ID)

    case fault.FAULT_T_TaskRunningErr:
        e.handleMotorStopFault("task running too long")

    default:
        fmt.Printf("Unknown fault type: %v\n", faultMsg.FaultType)
    }
}