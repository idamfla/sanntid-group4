package elevator

import "fmt"

func (e *Elevator) enterOfflineMode() {

    fmt.Println("Entering offline mode (cab-only)")
    e.offline = true

    for f := 0; f < len(e.hallRequests); f++ {
        elevio.SetButtonLamp(elevio.BT_HallUp, f, false)
        elevio.SetButtonLamp(elevio.BT_HallDown, f, false)
    }
}
}

func (e *Elevator) exitOfflineMode() {
    fmt.Println("Exiting offline mode (back online)")
    e.offline = false
}

func (e *Elevator) handleMotorStopFault(reason string){
   //TODO: implement
}

func (e *Elevator) handleNetworkFault(reason string) {
    fmt.Println("FAULT:", reason)
    //TODO: Check which fault
    e.offline = true
    elevio.SetMotorDirection(elevio.MD_Stop)

    e.elevatorState= ES_Idle
    e.faultTolerance.restartSelf()  // hvis dere vil auto-restart
}

func (e *Elevator) handlePeerDead(peerID int) {
    fmt.Println("Peer dead:", peerID)
    f !e.isMaster {
        return
    }

    // TODO senere: reassign hall calls
    // Midlertidig: marker peer dead i elevatorRegistry / fjern den
}


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
    e.faultTolerance.SetRoleMaster()
}

func (e *Elevator) BecameSlave() {
    e.isMaster = false
    e.faultTolerance.SetRoleSlave()
}