package server

import (
	"sync"
)

type ServerState struct {
	IsMaster           bool
	SearchingForMaster bool
	Synced             bool
	mu                 sync.Mutex
}

func (ss *ServerState) Reset() {
	ss.lock()
	defer ss.unlock()

	ss.IsMaster = false
	ss.Synced = false
	ss.SearchingForMaster = false
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
	ss.SearchingForMaster = false
}

func (ss *ServerState) SetMasterSearch() {
	ss.lock()
	defer ss.unlock()
	ss.SearchingForMaster = true
}

func (ss *ServerState) IsSearchingForMaster() bool {
	ss.lock()
	defer ss.unlock()
	return ss.SearchingForMaster
}

func (ss *ServerState) ClearMasterSearch() {
	ss.lock()
	defer ss.unlock()
	ss.SearchingForMaster = false
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
