package server

import (
	"fmt"
	"sync"
)

type PeerManager struct {
	peers map[string]*PeerInfo
	mu    sync.Mutex
}

func NewPeerManager() *PeerManager {
	return &PeerManager{peers: make(map[string]*PeerInfo)}
}

func (pm *PeerManager) AddPeer(peerKey string, p *PeerInfo) {
	pm.lock()
	defer pm.unlock()
	p.SetActiveNow()
	pm.peers[peerKey] = p
}

func (pm *PeerManager) CountAlivePeers() int {
	peers := pm.SnapshotPeers()

	count := 0
	for _, p := range peers {
		_, _, _, active, _, _ := p.Snapshot()
		if active {
			count++
		}
	}
	return count
}

func (pm *PeerManager) ClearAllAlive() {
	peers := pm.SnapshotPeers()

	for i, peer := range peers {
		if peer == nil {
			fmt.Printf("%s -> nil\n", i)
			continue
		}

		peer.ClearAlive()
	}
}

func (pm *PeerManager) GetPeer(peerKey string) (*PeerInfo, bool) {
	pm.lock()
	defer pm.unlock()
	peer, exists := pm.peers[peerKey]
	return peer, exists
}

func (pm *PeerManager) GetMasterPeer() *PeerInfo {
	peers := pm.SnapshotPeers()

	for _, p := range peers {
		_, _, isMaster, _, _, _ := p.Snapshot()
		if isMaster {
			return p
		}
	}
	return nil
}

// when called by server, trigger srv.clearMasterSearch()
func (pm *PeerManager) SetMasterPeer(peerKey string) {
	if peer, exists := pm.GetPeer(peerKey); exists {
		peer.SetMaster()
	}
}

func (pm *PeerManager) ClearMasterPeer() {
	if mstr := pm.GetMasterPeer(); mstr != nil {
		mstr.ClearMaster()
	}
}

func (pm *PeerManager) SetSynced(peerKey string) {
	if peer, exists := pm.GetPeer(peerKey); exists {
		peer.SetSynced()
	}
}

func (pm *PeerManager) ClearSynced(peerKey string) { // TODO is this used?
	if peer, exists := pm.GetPeer(peerKey); exists {
		peer.ClearSynced()
	}
}

func (pm *PeerManager) SetAliveNow(peerKey string) {
	if peer, exists := pm.GetPeer(peerKey); exists {
		peer.SetActiveNow()
	}
}

func (pm *PeerManager) SnapshotPeers() map[string]*PeerInfo {
	pm.lock()
	defer pm.unlock()

	peers := make(map[string]*PeerInfo, len(pm.peers))
	for id, p := range pm.peers {
		peers[id] = p
	}
	return peers
}

func (pm *PeerManager) PrintPeers() {
	peers := pm.SnapshotPeers()

	msg := "=== Peers map ===\n"

	for i, peer := range peers {
		if peer == nil {

			msg += fmt.Sprintf("%s -> nil\n", i)
			continue
		}
		id, addr, isMaster, active, synced, lastSeen := peer.Snapshot()

		msg += fmt.Sprintf(`%s -> Addr: %v
	IsMaster: %v
	Alive:   %v
	IsSynced: %v
	LastSeen: %v
`,
			id, addr, isMaster, active, synced, lastSeen)
	}

	msg += `=================
	`
	fmt.Println(msg)
}

func (pm *PeerManager) lock()   { pm.mu.Lock() }
func (pm *PeerManager) unlock() { pm.mu.Unlock() }
