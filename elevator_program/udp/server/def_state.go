package server

import (
	"sync"
)

type ServerState struct {
	IsMaster bool
	Synced   bool
	mu       sync.Mutex
}

func (ss *ServerState) Reset() {
	ss.lock()
	defer ss.unlock()

	ss.IsMaster = false
	ss.Synced = false
}

func (ss *ServerState) GetIsMaster() bool {
	ss.lock()
	defer ss.unlock()
	return ss.IsMaster
}

func (ss *ServerState) SetMaster() {
	ss.lock()
	defer ss.unlock()

	ss.IsMaster = true
	ss.Synced = true
}

func (ss *ServerState) ClearMaster() {
	ss.lock()
	defer ss.unlock()

	ss.IsMaster = false
}

func (ss *ServerState) GetSynced() bool {
	ss.lock()
	defer ss.unlock()
	return ss.Synced
}

func (ss *ServerState) SetSynced() {
	ss.lock()
	defer ss.unlock()
	ss.Synced = true
}

func (ss *ServerState) ClearSynced() {
	ss.lock()
	defer ss.unlock()
	ss.Synced = false
}

func (ss *ServerState) lock()   { ss.mu.Lock() }
func (ss *ServerState) unlock() { ss.mu.Unlock() }
