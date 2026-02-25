package udp

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	ID       string
	recvConn *net.UDPConn
	sendConn *net.UDPConn
	sessions map[uint32]*Session
	mu       sync.Mutex
	closeReq chan uint32

	done chan struct{}
	wg   sync.WaitGroup
}

func NewServer(ip string, port int, id string) (*Server, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip), // parse the string IP
		Port: port,
	}

	// make sockets
	recvConn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, err
	}

	// create a local UDP socket for sending (unconnected)
	sendAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"), // binds to any local IP
		Port: 0,                      // 0 = let OS pick a free port
	}

	sendConn, err := net.ListenUDP("udp", sendAddr)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		ID:       id,
		recvConn: recvConn,
		sendConn: sendConn,
		sessions: make(map[uint32]*Session),
		closeReq: make(chan uint32),
		done:     make(chan struct{}),
	}

	return srv, nil
}

func (srv *Server) Close() {
	close(srv.done)      // signal shutdown
	srv.recvConn.Close() // unblock ReadFromUDP
	srv.sendConn.Close()
	srv.wg.Wait() // wait for goroutines

	srv.mu.Lock()
	for id := range srv.sessions {
		srv.closeSessionLocked(id)
	}
	srv.mu.Unlock()

	close(srv.closeReq)
}

// helper function, not called directly
func (srv *Server) closeSessionLocked(sesID uint32) {
	ses, exists := srv.sessions[sesID]
	if exists {
		ses.Close()
		delete(srv.sessions, sesID)

		// TODO remove db
		fmt.Printf("Server %s removed session: %d\n", srv.ID, sesID)

	}
}

func (srv *Server) CloseSession(sesID uint32) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	srv.closeSessionLocked(sesID)
}

func (srv *Server) readLoop(out chan<- incommingPacket) {
	defer srv.wg.Done()
	buf := make([]byte, 2048)

	for {
		select {
		case <-srv.done:
			return
		default:
		}

		n, addr, err := srv.recvConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Read error:", err)
			continue
		}

		pck := decodePacket(buf, n)

		out <- incommingPacket{
			packet: pck,
			addr:   addr,
		}
	}
}

func (srv *Server) handleIncoming(incPck incommingPacket) {
	id := incPck.packet.Header.SessionID

	ses, exists := srv.sessions[id]
	if !exists {
		ses = NewSession(id, incPck.addr, srv.closeReq, srv)
		srv.sessions[id] = ses
	}

	ses.incoming <- incPck
}

func (srv *Server) Listen() {
	packetChan := make(chan incommingPacket)

	// UDP reader goroutine
	srv.wg.Add(1)
	go srv.readLoop(packetChan)

	// Main event loop
	for {
		select {
		case <-srv.done:
			// srv.Close()
			return

		case id := <-srv.closeReq:
			srv.CloseSession(id)

		case inc := <-packetChan:
			srv.handleIncoming(inc)
		}
	}
}

// TODO dont send string but rather the Message-struct
func (srv *Server) SendMessage(remoteAddr *net.UDPAddr, seq uint32, sessionID uint32, msg string) error {
	localAddr := srv.recvConn.LocalAddr().(*net.UDPAddr)

	dataPacket := Packet{
		Header: Header{
			Seq:           seq,
			MsgType:       MSG_T_Data,
			SessionID:     sessionID,
			SenderAddr:    localAddr.String(),
			RecipientAddr: remoteAddr.String(),
		},
		Payload: Message{Content: msg},
	}

	return sendPacket(srv.sendConn, remoteAddr, dataPacket)
}

func (srv *Server) SendReply(remoteAddr *net.UDPAddr, pck Packet, msgType MessageType) error {
	h := pck.Header

	replyContent := ""
	switch msgType {
	case MSG_T_Ack:
		replyContent = "ACK"
	case MSG_T_Commit:
		replyContent = "COMMIT"
	case MSG_T_Done:
		replyContent = "DONE"
	}

	reply := Packet{
		Header: Header{
			Seq:           h.Seq + 1,
			MsgType:       msgType,
			SessionID:     h.SessionID,
			SenderAddr:    h.RecipientAddr,
			RecipientAddr: remoteAddr.String(),
		},
		Payload: Message{Content: replyContent},
	}
	return sendPacket(srv.sendConn, remoteAddr, reply)
}

func (srv *Server) PrintSessions() {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	fmt.Printf("Active sessions (%d):\n", len(srv.sessions))
	// for id := range srv.sessions {
	// 	fmt.Println(" -", id)
	// }
}
