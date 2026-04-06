package server

import (
	"elevator_program/udp/packet"
	"fmt"
	"net"
)

func (srv *Server) updatePeer(peerAddr *net.UDPAddr, origin packet.Identity) {
	key := peerAddr.String()
	var isNew bool

	peer, exists := srv.getPeer(key)
	if !exists {
		if key != origin.Identifier {
			fmt.Println("Got a msg from a unknown peer that is not itself the origin of this message ... whoIsAlive?")
			srv.QueueWhoIsAliveMsg()
			return
		}

		peer = NewPeer(origin.Alias, peerAddr)

		srv.addPeer(key, peer)
		fmt.Printf("Server %s: new peer made: %s\n", srv.GetAlias(), key)
		isNew = true
	}

	srv.setActiveNow(key)

	wasRevived := !isNew && !peer.Active
	if wasRevived {
		fmt.Println("This peer was revived ... do something maybe?") // TODO syncing
	}
}

func (srv *Server) getPeer(peerKey string) (*PeerInfo, bool) { return srv.peers.getPeer(peerKey) }
func (srv *Server) addPeer(peerKey string, p *PeerInfo)      { srv.peers.addPeer(peerKey, p) }
func (srv *Server) setActiveNow(peerKey string)              { srv.peers.setActiveNow(peerKey) }
func (srv *Server) countActivePeers() int                    { return srv.peers.countActivePeers() }
func (srv *Server) getMasterPeer() *PeerInfo                 { return srv.peers.getMasterPeer() }
func (srv *Server) setMasterPeer(peerKey string)             { srv.peers.setMasterPeer(peerKey) }
func (srv *Server) clearMasterPeer()                         { srv.peers.clearMasterPeer() }
func (srv *Server) PrintPeers()                              { srv.peers.PrintPeers() }
