package server

import (
	"elevator_program/message"
	"fmt"
	"net"
	"sync"
	"time"
)

type PeerInfo struct {
	Alias     string
	Addr      *net.UDPAddr
	LastSeen  time.Time
	IsSynced  bool
	Active    bool
	IsMaster  bool
	EMsgQueue chan message.ElevatorMessage // TODO might fade out
	mu        sync.Mutex
}

func NewPeer(alias string, addr *net.UDPAddr) *PeerInfo {
	return &PeerInfo{
		Alias:     alias,
		Addr:      addr,
		LastSeen:  time.Now(),
		IsSynced:  false,
		Active:    true,
		IsMaster:  false,
		EMsgQueue: make(chan message.ElevatorMessage, CHANNEL_BUF),
	}
}

func (p *PeerInfo) GetAddr() *net.UDPAddr { return p.Addr }
func (p *PeerInfo) GetAddrString() string { return p.GetAddr().String() }

func (p *PeerInfo) IsActive() bool {
	p.lock()
	defer p.unlock()
	return p.IsActive()
}

func (p *PeerInfo) SetActiveNow() {
	p.lock()
	defer p.unlock()
	p.Active = true
	p.LastSeen = time.Now()
}

func (p *PeerInfo) ClearActive() {
	p.lock()
	defer p.unlock()
	p.Active = false
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

func (p *PeerInfo) QueueMessage(msg message.ElevatorMessage) {
	select {
	case p.EMsgQueue <- msg:
	default:
		fmt.Println("p EMsgQueue is full ... dropping packet")
	}
}

func (p *PeerInfo) Snapshot() (alias string, addr *net.UDPAddr, isMaster, active, synced bool, lastSeen time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Alias, p.Addr, p.IsMaster, p.Active, p.IsSynced, p.LastSeen
}

func (p *PeerInfo) lock()   { p.mu.Lock() }
func (p *PeerInfo) unlock() { p.mu.Unlock() }
