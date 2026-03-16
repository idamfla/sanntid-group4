package elevator

import (
	"elevator_program/fault"
	"testing"
	"time"
)

func newFailoverTestElevator(id int) *Elevator {
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

func TestElectionFailover_LowerIDJoinsThenDies(t *testing.T) {
	e := newFailoverTestElevator(3)

	// 1) Elevator 3 starts alone -> should become master
	e.runElection("startup")

	if !e.isMaster {
		t.Fatal("expected elevator 3 to become master when alone")
	}
	if e.currentMasterID != 3 {
		t.Fatalf("expected currentMasterID = 3, got %d", e.currentMasterID)
	}

	// 2) Elevator 1 appears -> lowest ID should win, so 3 becomes slave
	e.ObservePeer(1)

	if e.isMaster {
		t.Fatal("expected elevator 3 to step down when peer 1 appears")
	}
	if e.currentMasterID != 1 {
		t.Fatalf("expected currentMasterID = 1 after peer 1 appears, got %d", e.currentMasterID)
	}

	// 3) Elevator 1 dies -> remove peer and run election again
	e.faultTolerance.RemovePeer(1)
	e.runElection("peer 1 died")

	if !e.isMaster {
		t.Fatal("expected elevator 3 to become master again after peer 1 dies")
	}
	if e.currentMasterID != 3 {
		t.Fatalf("expected currentMasterID = 3 after failover, got %d", e.currentMasterID)
	}
}

func TestElectionFailover_NextLowestAliveWins(t *testing.T) {
	e := newFailoverTestElevator(4)

	// Elevator 4 sees three peers
	e.faultTolerance.SeenPeer(2)
	e.faultTolerance.SeenPeer(3)
	e.faultTolerance.SeenPeer(5)

	e.runElection("initial election")

	if e.isMaster {
		t.Fatal("expected elevator 4 not to be master when peer 2 exists")
	}
	if e.currentMasterID != 2 {
		t.Fatalf("expected master 2, got %d", e.currentMasterID)
	}

	// Lowest one dies -> next lowest alive should win
	e.faultTolerance.RemovePeer(2)
	e.runElection("peer 2 died")

	if e.isMaster {
		t.Fatal("expected elevator 4 still not to be master when peer 3 exists")
	}
	if e.currentMasterID != 3 {
		t.Fatalf("expected master 3 after peer 2 dies, got %d", e.currentMasterID)
	}

	// Next one dies too -> now elevator 4 should become master
	e.faultTolerance.RemovePeer(3)
	e.runElection("peer 3 died")

	if !e.isMaster {
		t.Fatal("expected elevator 4 to become master when 2 and 3 are gone")
	}
	if e.currentMasterID != 4 {
		t.Fatalf("expected master 4 after failover, got %d", e.currentMasterID)
	}
}


func TestHandlePeerDead_TriggersFailoverWhenMasterDies(t *testing.T) {
	e := newFailoverTestElevator(3)

	// Start with peer 1 alive, so elevator 3 becomes slave under 1
	e.faultTolerance.SeenPeer(1)
	e.runElection("initial")

	if e.isMaster {
		t.Fatal("expected elevator 3 to start as slave when peer 1 is alive")
	}
	if e.currentMasterID != 1 {
		t.Fatalf("expected currentMasterID = 1, got %d", e.currentMasterID)
	}

	// Simulate that peer 1 dies
	e.handlePeerDead(1)

	if !e.isMaster {
		t.Fatal("expected elevator 3 to become master after master peer dies")
	}
	if e.currentMasterID != 3 {
		t.Fatalf("expected currentMasterID = 3 after failover, got %d", e.currentMasterID)
	}
}

