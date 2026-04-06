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

func (pm *PeerManager) addPeer(peerKey string, p *PeerInfo) {
	pm.lock()
	defer pm.unlock()
	p.SetActiveNow()
	pm.peers[peerKey] = p
}

func (pm *PeerManager) countAlivePeers() int {
	peers := pm.snapshotPeers()

	count := 0
	for _, p := range peers {
		_, _, _, active, _, _ := p.Snapshot()
		if active {
			count++
		}
	}
	return count
}

func (pm *PeerManager) clearAllAlive() {
	peers := pm.snapshotPeers()

	for i, peer := range peers {
		if peer == nil {
			fmt.Printf("%s -> nil\n", i)
			continue
		}

		peer.ClearAlive()
	}
}

func (pm *PeerManager) getPeer(peerKey string) (*PeerInfo, bool) {
	pm.lock()
	defer pm.unlock()
	peer, exists := pm.peers[peerKey]
	return peer, exists
}

func (pm *PeerManager) getMasterPeer() *PeerInfo {
	peers := pm.snapshotPeers()

	for _, p := range peers {
		_, _, isMaster, _, _, _ := p.Snapshot()
		if isMaster {
			return p
		}
	}
	return nil
}

// when called by server, trigger srv.clearMasterSearch()
func (pm *PeerManager) setMasterPeer(peerKey string) {
	if peer, exists := pm.getPeer(peerKey); exists {
		peer.SetMaster()
	}
}

func (pm *PeerManager) clearMasterPeer() {
	if mstr := pm.getMasterPeer(); mstr != nil {
		mstr.ClearMaster()
	}
}

func (pm *PeerManager) setSynced(peerKey string) {
	if peer, exists := pm.getPeer(peerKey); exists {
		peer.SetSynced()
	}
}

func (pm *PeerManager) clearSynced(peerKey string) {
	if peer, exists := pm.getPeer(peerKey); exists {
		peer.ClearSynced()
	}
}

func (pm *PeerManager) setAliveNow(peerKey string) {
	if peer, exists := pm.getPeer(peerKey); exists {
		peer.SetActiveNow()
	}
}

func (pm *PeerManager) snapshotPeers() map[string]*PeerInfo {
	pm.lock()
	defer pm.unlock()

	peers := make(map[string]*PeerInfo, len(pm.peers))
	for id, p := range pm.peers {
		peers[id] = p
	}
	return peers
}

func (pm *PeerManager) PrintPeers() {
	peers := pm.snapshotPeers()

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
