package fault

import (
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	mu sync.Mutex

	cfg  Config
	role Role
	id   string

	startedAt time.Time

	lastSeenMaster time.Time
	lastSeenPeer   map[string]time.Time

	lastFloorEvent time.Time
	motorRunning   bool

	online bool
	faulty bool

	OnBecomeMaster    func()
	OnMasterSuspected func(reason string)
	OnPeerDead        func(peerID string)
	OnGoOnline        func()
	OnGoOffline       func()
	OnMotorFault      func(reason string)
	OnNetworkFault    func(reason string)
}

func NewFaultManager(id string, cfg Config) *Manager {
	return &Manager{
		cfg:            cfg,
		id:             id,
		role:           RoleSlave,
		online:         true,
		startedAt:      time.Now(),
		lastSeenPeer:   make(map[string]time.Time),
		lastFloorEvent: time.Now(),
		lastSeenMaster: time.Now(),
	}
}

func (fm *Manager) SeenMaster() {
	fm.mu.Lock()

	fm.lastSeenMaster = time.Now()

	var goOnline func()
	if !fm.online {
		fm.online = true
		goOnline = fm.OnGoOnline
	}

	fm.mu.Unlock()

	if goOnline != nil {
		goOnline()
	}
}

func (fm *Manager) SeenPeer(peerID string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.lastSeenPeer[peerID] = time.Now()
}

func (fm *Manager) RemovePeer(peerID string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	delete(fm.lastSeenPeer, peerID)
}

func (fm *Manager) FloorEvent() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.lastFloorEvent = time.Now()
}

func (fm *Manager) SetMotorRunning(running bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.motorRunning = running
	if running {
		fm.lastFloorEvent = time.Now()
		fm.faulty = false
	}
}

func (fm *Manager) SetRoleMaster() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.role = RoleMaster
	fm.lastSeenMaster = time.Now()
	fm.online = true
}

func (fm *Manager) SetRoleSlave() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.role = RoleSlave
}

func (fm *Manager) Role() Role {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	return fm.role
}

func (fm *Manager) AlivePeers() []string {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	now := time.Now()
	alive := make([]string, 0, len(fm.lastSeenPeer))

	for peerID, ts := range fm.lastSeenPeer {
		if now.Sub(ts) <= fm.cfg.PeerTimeout {
			alive = append(alive, peerID)
		}
	}

	return alive
}

func (fm *Manager) checkMasterTimeout() {
	fm.mu.Lock()

	if time.Since(fm.startedAt) < fm.cfg.StartupGrace {
		fm.mu.Unlock()
		return
	}

	if fm.role != RoleSlave {
		fm.mu.Unlock()
		return
	}

	if fm.lastSeenMaster.IsZero() {
		fm.mu.Unlock()
		return
	}

	var onMasterSuspected func(string)
	var onGoOffline func()
	var onNetworkFault func(string)
	shouldNotify := false

	if time.Since(fm.lastSeenMaster) > fm.cfg.MasterTimeout {
		fmt.Println("Master timeout detected")

		if fm.online {
			fm.online = false
			shouldNotify = true
			onMasterSuspected = fm.OnMasterSuspected
		}
	}

	fm.mu.Unlock()

	if shouldNotify {
		if onMasterSuspected != nil {
			onMasterSuspected("master timeout")
		}
	}

	if onGoOffline != nil {
		onGoOffline()
	}
	if onNetworkFault != nil {
		onNetworkFault("master timeout")
	}
}

func (fm *Manager) checkPeerTimeout() {
	fm.mu.Lock()

	if fm.role != RoleMaster {
		fm.mu.Unlock()
		return
	}

	now := time.Now()
	deadPeers := make([]string, 0)

	for peerID, ts := range fm.lastSeenPeer {
		if now.Sub(ts) > fm.cfg.PeerTimeout {
			fmt.Println("Peer timeout:", peerID)
			delete(fm.lastSeenPeer, peerID)
			deadPeers = append(deadPeers, peerID)
		}
	}

	onPeerDead := fm.OnPeerDead

	fm.mu.Unlock()

	if onPeerDead != nil {
		for _, peerID := range deadPeers {
			onPeerDead(peerID)
		}
	}
}

func (fm *Manager) checkMotorTimeout() {
	fm.mu.Lock()

	if !fm.motorRunning {
		fm.mu.Unlock()
		return
	}

	var onMotorFault func(string)
	shouldNotify := false

	if time.Since(fm.lastFloorEvent) > fm.cfg.MotorTimeout {
		if !fm.faulty {
			fm.faulty = true
			shouldNotify = true
			onMotorFault = fm.OnMotorFault
		}
	}

	fm.mu.Unlock()

	if shouldNotify && onMotorFault != nil {
		onMotorFault("motor watchdog timeout")
	}
}

func (fm *Manager) Run() {
	ticker := time.NewTicker(fm.cfg.Tick)
	defer ticker.Stop()

	for range ticker.C {
		fm.checkMasterTimeout()
		fm.checkPeerTimeout()
		fm.checkMotorTimeout()
	}
}


