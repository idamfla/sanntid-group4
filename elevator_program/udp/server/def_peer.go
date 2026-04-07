package server

import (
	"net"
	"sync"
	"time"
)

type PeerInfo struct {
	Alias    string
	Addr     *net.UDPAddr
	LastSeen time.Time
	IsSynced bool
	Alive    bool
	IsMaster bool
	mu       sync.Mutex
}

func NewPeer(alias string, addr *net.UDPAddr) *PeerInfo {
	return &PeerInfo{
		Alias:    alias,
		Addr:     addr,
		LastSeen: time.Now(),
		IsSynced: false,
		Alive:    true,
		IsMaster: false,
	}
}

func (p *PeerInfo) GetAddr() *net.UDPAddr { return p.Addr }
func (p *PeerInfo) GetAddrString() string { return p.GetAddr().String() }

func (p *PeerInfo) IsAlive() bool {
	p.lock()
	defer p.unlock()
	return p.Alive
}

func (p *PeerInfo) SetActiveNow() {
	p.lock()
	defer p.unlock()
	p.Alive = true
	p.LastSeen = time.Now()
}

func (p *PeerInfo) ClearAlive() {
	p.lock()
	defer p.unlock()
	p.Alive = false
}

func (p *PeerInfo) SetMaster() {
	p.lock()
	defer p.unlock()
	p.IsMaster = true
}
func (p *PeerInfo) ClearMaster() {
	p.lock()
	defer p.unlock()
	p.IsMaster = false
}

func (p *PeerInfo) SetSynced() {
	p.lock()
	defer p.unlock()
	p.IsSynced = true
}
func (p *PeerInfo) ClearSynced() {
	p.lock()
	defer p.unlock()
	p.IsSynced = false
}

func (p *PeerInfo) Snapshot() (alias string, addr *net.UDPAddr, isMaster, active, synced bool, lastSeen time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Alias, p.Addr, p.IsMaster, p.Alive, p.IsSynced, p.LastSeen
}

func (p *PeerInfo) lock()   { p.mu.Lock() }
func (p *PeerInfo) unlock() { p.mu.Unlock() }
