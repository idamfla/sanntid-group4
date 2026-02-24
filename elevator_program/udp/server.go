package udp

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	conn     *net.UDPConn
	sessions map[uint32]*Session
	mu       sync.Mutex
	closeReq chan uint32
}

func NewServer(ip string, port int) (*Server, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip), // parse the string IP
		Port: port,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		conn:     conn,
		sessions: make(map[uint32]*Session),
		closeReq: make(chan uint32),
	}

	return srv, nil
}

func (srv *Server) CloseSession(sesID uint32) {
	srv.mu.Lock()
	ses, exists := srv.sessions[sesID]
	if exists {
		ses.Close()
		delete(srv.sessions, sesID)
	}
	srv.mu.Unlock()
}

func (srv *Server) readLoop(out chan<- incomingPacket) {
	buf := make([]byte, 2048)

	for {
		n, addr, err := srv.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Read error:", err)
			continue
		}

		pck := decodePacket(buf, n)

		out <- incomingPacket{
			packet: pck,
			addr:   addr,
		}
	}
}

func (srv *Server) handleIncoming(inc incomingPacket) {
	id := inc.packet.Header.SessionID

	ses, exists := srv.sessions[id]
	if !exists {
		ses = NewSession(id, inc.addr, srv.conn, srv.closeReq)
		srv.sessions[id] = ses
	}

	ses.incoming <- inc.packet
}

func (srv *Server) Listen() {
	packetChan := make(chan incomingPacket)

	// UDP reader goroutine
	go srv.readLoop(packetChan)

	// Main event loop
	for {
		select {

		case id := <-srv.closeReq:
			srv.CloseSession(id)
			fmt.Println("Server removed session:", id)

		case inc := <-packetChan:
			srv.handleIncoming(inc)
		}
	}
}
