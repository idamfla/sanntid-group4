package server

import "net"

func (srv *Server) getRecvConn() *net.UDPConn {
	return srv.network.GetRecvConn()
}

func (srv *Server) getRecvString() string {
	return srv.getRecvAddr().String()
}

func (srv *Server) getRecvUDPAddr() *net.UDPAddr {
	return srv.getRecvAddr().(*net.UDPAddr)
}

func (srv *Server) getRecvAddr() net.Addr {
	return srv.network.GetRecvConn().LocalAddr()
}

func (srv *Server) getSendConn() *net.UDPConn {
	return srv.network.GetSendConn()
}

func (srv *Server) getBroadcastConn() *net.UDPConn {
	return srv.network.GetBroadcastConn()
}

func (srv *Server) getBroadcastAddr() *net.UDPAddr {
	return srv.network.GetBroadcastAddr()
}
