package elevator

import (
	"elevator_program/fault"
	"testing"
	"time"
)

func newSmallTestElevator(id int) *Elevator {
	e := &Elevator{}
	e.id = id
	e.currentMasterID = -1
	e.system.Elevators = make(map[int]ElevatorsStatus)
	e.elevatorRegistry = make(map[int]ElevatorsStatus)

	e.faultTolerance = fault.NewFaultManager(id, fault.Config{
		StartupGrace:  20 * time.Millisecond,
		MasterTimeout: 30 * time.Millisecond,
		PeerTimeout:   50 * time.Millisecond,
		MotorTimeout:  50 * time.Millisecond,
		Tick:          5 * time.Millisecond,
	})

	return e
}

func TestChooseMasterID_ReturnsOwnIDWhenAlone(t *testing.T) {
	e := newSmallTestElevator(3)

	got := e.chooseMasterID()

	if got != 3 {
		t.Fatalf("expected master id 3, got %d", got)
	}
}

func TestChooseMasterID_ReturnsLowestAliveID(t *testing.T) {
	e := newSmallTestElevator(3)

	e.faultTolerance.SeenPeer(5)
	e.faultTolerance.SeenPeer(2)
	e.faultTolerance.SeenPeer(7)

	got := e.chooseMasterID()

	if got != 2 {
		t.Fatalf("expected lowest id 2, got %d", got)
	}
}

func TestRunElection_BecomesMasterWhenAlone(t *testing.T) {
	e := newSmallTestElevator(3)

	e.runElection("test alone")

	if !e.isMaster {
		t.Fatal("expected elevator to become master")
	}

	if e.currentMasterID != 3 {
		t.Fatalf("expected currentMasterID 3, got %d", e.currentMasterID)
	}

	if !e.connectedToMaster {
		t.Fatal("expected connectedToMaster to be true")
	}

	if e.faultTolerance.Role() != fault.RoleMaster {
		t.Fatalf("expected fault manager role master, got %v", e.faultTolerance.Role())
	}
}

func TestRunElection_BecomesSlaveWhenLowerPeerExists(t *testing.T) {
	e := newSmallTestElevator(3)
	e.faultTolerance.SeenPeer(2)

	e.runElection("test lower peer exists")

	if e.isMaster {
		t.Fatal("expected elevator not to be master")
	}

	if e.currentMasterID != 2 {
		t.Fatalf("expected currentMasterID 2, got %d", e.currentMasterID)
	}

	if !e.connectedToMaster {
		t.Fatal("expected connectedToMaster to be true")
	}

	if e.faultTolerance.Role() != fault.RoleSlave {
		t.Fatalf("expected fault manager role slave, got %v", e.faultTolerance.Role())
	}
}

func TestObservePeer_ChoosesLowerPeerAsMaster(t *testing.T) {
	e := newSmallTestElevator(4)

	e.ObservePeer(2)

	if e.currentMasterID != 2 {
		t.Fatalf("expected currentMasterID 2, got %d", e.currentMasterID)
	}

	if e.isMaster {
		t.Fatal("expected elevator not to be master")
	}

	if !e.connectedToMaster {
		t.Fatal("expected connectedToMaster to be true")
	}
}

func TestObserveMaster_MakesMasterStepDownForLowerID(t *testing.T) {
	e := newSmallTestElevator(4)

	e.isMaster = true
	e.currentMasterID = 4
	e.connectedToMaster = true
	e.faultTolerance.SetRoleMaster()

	e.ObserveMaster(2)

	if e.isMaster {
		t.Fatal("expected elevator to step down and become slave")
	}

	if e.currentMasterID != 2 {
		t.Fatalf("expected currentMasterID 2, got %d", e.currentMasterID)
	}

	if !e.connectedToMaster {
		t.Fatal("expected connectedToMaster to stay true")
	}

	if e.faultTolerance.Role() != fault.RoleSlave {
		t.Fatalf("expected fault manager role slave, got %v", e.faultTolerance.Role())
	}
}