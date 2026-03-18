package elevator

import (
	"fmt"
	"sort"
)

func (e *Elevator) chooseMasterID() string {

    // Start med egen ID som kandidat
    candidates := []string{e.Id}

    // Legg til alle levende peers hvis faultTolerance finnes
    if e.faultTolerance != nil {
        candidates = append(candidates, e.faultTolerance.AlivePeers()...)
    }

    // Sorter kandidatene for å velge den minste
    sort.Strings(candidates)

	return candidates[0]
}

func (e *Elevator) runElection(reason string) {
	chosenMaster := e.chooseMasterID()

	fmt.Printf("Elevator %s runs election (%s). Chosen master: %s\n", e.Id, reason, chosenMaster)

	if chosenMaster == e.Id {
		if !e.IsMaster {
			e.BecameMaster()
		} else {
			e.currentMasterID = e.Id
			e.connectedToMaster = true
		}
		return
	}

	if e.IsMaster || e.currentMasterID != chosenMaster {
		e.BecameSlave(chosenMaster)
	}
}

func (e *Elevator) BecameMaster() {
	fmt.Printf("Elevator %s became MASTER\n", e.Id)

	e.IsMaster = true
	e.IsOnline = true
	e.currentMasterID = e.Id
	e.connectedToMaster = true

	if e.faultTolerance != nil {
		e.faultTolerance.SetRoleMaster()
	}

	e.InitMasterElevator()
	go e.RunMasterLoop()
}

func (e *Elevator) BecameSlave(masterID string) {
	fmt.Printf("Elevator %s became SLAVE. Master is %s\n", e.Id, masterID)

	e.IsMaster = false
	e.IsOnline = true
	e.currentMasterID = masterID
	e.connectedToMaster = true

	if e.faultTolerance != nil {
		e.faultTolerance.SetRoleSlave()
	}
}

func (e *Elevator) ObservePeer(peerID string) {
	if peerID == e.Id {
		return
	}

	if e.faultTolerance != nil {
		e.faultTolerance.SeenPeer(peerID)
	}

	if e.currentMasterID == "-1" || peerID < e.currentMasterID {
		e.runElection("peer observed")
	}
}

func (e *Elevator) ObserveMaster(masterID string) {
	if masterID == e.Id {
		return
	}

	if e.faultTolerance != nil {
		e.faultTolerance.SeenMaster()
		e.faultTolerance.SeenPeer(masterID)
	}

	e.currentMasterID = masterID
	e.connectedToMaster = true
	e.IsOnline = true

	if e.IsMaster && masterID < e.Id {
		e.BecameSlave(masterID)
	}
}
