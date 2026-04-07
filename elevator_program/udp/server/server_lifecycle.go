package server

import (
	"sync"
)

type ServerLifecycle struct {
	CloseReq  chan uint32 // make the server/owner close this session
	Stop      chan struct{}
	Wg        sync.WaitGroup
	CloseOnce sync.Once
}

func NewServerLifecycle() *ServerLifecycle {
	return &ServerLifecycle{
		CloseReq: make(chan uint32, CHANNEL_BUF),
		Stop:     make(chan struct{}, 1),
	}
}

func (srv *Server) WgAdd(value int) { srv.lifecycle.Wg.Add(value) }
func (srv *Server) WgWait()         { srv.lifecycle.Wg.Wait() }
func (srv *Server) WgDone()         { srv.lifecycle.Wg.Done() }

func (srv *Server) CloseReqCh() chan uint32 { return srv.lifecycle.CloseReq }
func (srv *Server) stopCh() <-chan struct{} { return srv.lifecycle.Stop }
