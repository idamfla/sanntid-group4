package fault

import (
	"testing"
	"time"
)

func smallTestConfig() Config {
	return Config{
		StartupGrace:  20 * time.Millisecond,
		MasterTimeout: 30 * time.Millisecond,
		PeerTimeout:   30 * time.Millisecond,
		MotorTimeout:  30 * time.Millisecond,
		Tick:          5 * time.Millisecond,
	}
}

func TestNewFaultManager_StartsAsSlave(t *testing.T) {
	fm := NewFaultManager("1", smallTestConfig())

	if fm == nil {
		t.Fatal("expected fault manager, got nil")
	}

	if fm.Role() != RoleSlave {
		t.Fatalf("expected initial role RoleSlave, got %v", fm.Role())
	}

	if !fm.online {
		t.Fatal("expected manager to start online")
	}
}

func TestSeenPeer_AddsAlivePeer(t *testing.T) {
	fm := NewFaultManager("1", smallTestConfig())

	fm.SeenPeer("2")

	alive := fm.AlivePeers()

	if len(alive) != 1 {
		t.Fatalf("expected 1 alive peer, got %d", len(alive))
	}

	if alive[0] != "2" {
		t.Fatalf("expected peer 2 to be alive, got %d", alive[0])
	}
}

func TestAlivePeers_RemovesTimedOutPeer(t *testing.T) {
	fm := NewFaultManager("1", smallTestConfig())

	fm.SeenPeer("2")

	time.Sleep(fm.cfg.PeerTimeout + 10*time.Millisecond)

	alive := fm.AlivePeers()

	if len(alive) != 0 {
		t.Fatalf("expected 0 alive peers after timeout, got %d", len(alive))
	}
}

func TestMasterTimeout_TriggersMasterSuspected(t *testing.T) {
	fm := NewFaultManager("1", smallTestConfig())

	called := false
	gotReason := ""

	fm.OnMasterSuspected = func(reason string) {
		called = true
		gotReason = reason
	}

	go fm.Run()

	time.Sleep(fm.cfg.StartupGrace + fm.cfg.MasterTimeout + 20*time.Millisecond)

	if !called {
		t.Fatal("expected OnMasterSuspected to be called after master timeout")
	}

	if gotReason != "master timeout" {
		t.Fatalf("expected reason 'master timeout', got %q", gotReason)
	}
}

func TestSeenMaster_BringsNodeBackOnline(t *testing.T) {
	fm := NewFaultManager("1", smallTestConfig())

	wentOffline := false
	wentOnline := false

	fm.OnGoOffline = func() {
		wentOffline = true
	}

	fm.OnGoOnline = func() {
		wentOnline = true
	}

	go fm.Run()

	time.Sleep(fm.cfg.StartupGrace + fm.cfg.MasterTimeout + 20*time.Millisecond)

	if !wentOffline {
		t.Fatal("expected node to go offline after master timeout")
	}

	fm.SeenMaster()

	time.Sleep(10 * time.Millisecond)

	if !fm.online {
		t.Fatal("expected node to be online after SeenMaster")
	}

	if !wentOnline {
		t.Fatal("expected OnGoOnline to be called after SeenMaster")
	}
}

func TestPeerTimeout_CallsOnPeerDeadWhenMaster(t *testing.T) {
	fm := NewFaultManager("1", smallTestConfig())
	fm.SetRoleMaster()
	fm.SeenPeer("2")

	called := false
	deadID := "-1"

	fm.OnPeerDead = func(peerID string) {
		called = true
		deadID = peerID
	}

	go fm.Run()

	time.Sleep(fm.cfg.PeerTimeout + 20*time.Millisecond)

	if !called {
		t.Fatal("expected OnPeerDead to be called after peer timeout")
	}

	if deadID != "2" {
		t.Fatalf("expected dead peer id 2, got %d", deadID)
	}
}

func TestMotorTimeout_CallsOnMotorFault(t *testing.T) {
	fm := NewFaultManager("1", smallTestConfig())

	called := false
	gotReason := ""

	fm.OnMotorFault = func(reason string) {
		called = true
		gotReason = reason
	}

	fm.SetMotorRunning(true)

	go fm.Run()

	time.Sleep(fm.cfg.MotorTimeout + 20*time.Millisecond)

	if !called {
		t.Fatal("expected OnMotorFault to be called after motor timeout")
	}

	if gotReason != "motor watchdog timeout" {
		t.Fatalf("expected reason 'motor watchdog timeout', got %q", gotReason)
	}
}

func TestFloorEvent_PreventsMotorFault(t *testing.T) {
	fm := NewFaultManager("1", smallTestConfig())

	called := false

	fm.OnMotorFault = func(reason string) {
		called = true
	}

	fm.SetMotorRunning(true)
	go fm.Run()

	time.Sleep(10 * time.Millisecond)
	fm.FloorEvent()

	time.Sleep(10 * time.Millisecond)
	fm.FloorEvent()

	time.Sleep(10 * time.Millisecond)
	fm.FloorEvent()

	if called {
		t.Fatal("did not expect OnMotorFault while floor events keep arriving")
	}
}