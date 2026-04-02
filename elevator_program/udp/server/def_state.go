package server

import (
	"sync"
)

type ServerState struct {
	IsMaster           bool
	SearchingForMaster bool
	IsSynced           bool
	mu                 sync.Mutex
}

func (ss *ServerState) Reset() {
	ss.lock()
	defer ss.unlock()

	ss.IsMaster = false
	ss.IsSynced = false
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
	ss.IsSynced = true
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

func (ss *ServerState) GetIsSynced() bool {
	ss.lock()
	defer ss.unlock()
	return ss.IsSynced
}

func (ss *ServerState) SetIsSynced(isSynced bool) {
	ss.lock()
	defer ss.unlock()
	ss.IsSynced = isSynced
}

func (ss *ServerState) lock() {
	ss.mu.Lock()
}

func (ss *ServerState) unlock() {
	ss.mu.Unlock()
}
