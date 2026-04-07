package server

import (
	"elevator_program/message"
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

func (srv *Server) updatePeer(peerAddr *net.UDPAddr, origin packet.Identity) {
	key := peerAddr.String()
	peerAlias := origin.Alias
	var isNew bool

	peer, exists := srv.getPeer(key)
	if !exists {
		if key != origin.Identifier {
			fmt.Println("Got a msg from a unknown peer that is not itself the origin of this message ... whoIsAlive?")
			srv.QueueWhoIsAliveMsg()
			return
		}

		peer = NewPeer(peerAlias, peerAddr)

		srv.addPeer(key, peer)
		fmt.Printf("Server %s: new peer made: %s (aka. %s)\n", srv.GetAlias(), key, peerAlias)
		isNew = true
	}

	srv.setPeerAliveNow(key)

	wasRevived := !isNew && !peer.IsAlive()
	if wasRevived {
		fmt.Println("This peer was revived ... do something maybe?")
	}
}

func (srv *Server) getPeer(peerKey string) (*PeerInfo, bool) { return srv.peers.GetPeer(peerKey) }
func (srv *Server) addPeer(peerKey string, p *PeerInfo)      { srv.peers.AddPeer(peerKey, p) }

func (srv *Server) setPeerAliveNow(peerKey string) { srv.peers.SetAliveNow(peerKey) }
func (srv *Server) clearAllAlive() {
	srv.peers.ClearAllAlive()
	ready := make(chan struct{}, 1)
	select {
	case ready <- struct{}{}:
	default:
	}

	srv.QueueElevatorTask(message.ElevatorMessage{EMsgType: message.EMSG_T_IAmAlone}, nil, ready)
}
func (srv *Server) countAlivePeers() int { return srv.peers.CountAlivePeers() }

func (srv *Server) getMasterPeer() *PeerInfo { return srv.peers.GetMasterPeer() }
func (srv *Server) ClearMasterPeer()         { srv.peers.ClearMasterPeer() }

func (srv *Server) setPeerSynced(peerKey string) { srv.peers.SetSynced(peerKey) }
func (srv *Server) ClearAllPeerSynced()          { srv.peers.ClearAllSynced() }

func (srv *Server) PrintPeers() { srv.peers.PrintPeers() }
