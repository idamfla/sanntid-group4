package elevator

import (
	"fmt"
	"sort"
)

func (e *Elevator) chooseMasterID() int {
	candidates := []int{e.id}

	if e.faultTolerance != nil {
		candidates = append(candidates, e.faultTolerance.AlivePeers()...)
	}

	sort.Ints(candidates)
	return candidates[0]
}

func (e *Elevator) runElection(reason string) {
	chosenMaster := e.chooseMasterID()

	fmt.Printf("Elevator %d runs election (%s). Chosen master: %d\n", e.id, reason, chosenMaster)

	if chosenMaster == e.id {
		if !e.isMaster {
			e.BecameMaster()
		} else {
			e.currentMasterID = e.id
			e.connectedToMaster = true
		}
		return
	}

	if e.isMaster || e.currentMasterID != chosenMaster {
		e.BecameSlave(chosenMaster)
	}
}

func (e *Elevator) BecameMaster() {
	fmt.Printf("Elevator %d became MASTER\n", e.id)

	e.isMaster = true
	e.currentMasterID = e.id
	e.connectedToMaster = true

	if e.faultTolerance != nil {
		e.faultTolerance.SetRoleMaster()
	}

	e.InitMasterElevator()
	go e.RunMasterLoop()
}

func (e *Elevator) BecameSlave(masterID int) {
	fmt.Printf("Elevator %d became SLAVE. Master is %d\n", e.id, masterID)

	e.isMaster = false
	e.currentMasterID = masterID
	e.connectedToMaster = true

	if e.faultTolerance != nil {
		e.faultTolerance.SetRoleSlave()
	}
}

func (e *Elevator) ObservePeer(peerID int) {
	if peerID == e.id {
		return
	}

	if e.faultTolerance != nil {
		e.faultTolerance.SeenPeer(peerID)
	}

	if e.currentMasterID == -1 || peerID < e.currentMasterID {
		e.runElection("peer observed")
	}
}

func (e *Elevator) ObserveMaster(masterID int) {
	if masterID == e.id {
		return
	}

	if e.faultTolerance != nil {
		e.faultTolerance.SeenMaster()
		e.faultTolerance.SeenPeer(masterID)
	}

	e.currentMasterID = masterID
	e.connectedToMaster = true

	if e.isMaster && masterID < e.id {
		e.BecameSlave(masterID)
	}
}
