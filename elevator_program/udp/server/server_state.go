package server

func (srv *Server) ResetState() {
	srv.state.Reset()
}

func (srv *Server) IsMaster() bool     { return srv.state.GetIsMaster() }
func (srv *Server) setSelfAsMaster()   { srv.state.SetMaster() }
func (srv *Server) clearSelfAsMaster() { srv.state.ClearMaster() }

func (srv *Server) isSynced() bool { return srv.state.GetSynced() }
func (srv *Server) setSynced()     { srv.state.SetSynced() }
func (srv *Server) clearSynced()   { srv.state.ClearSynced() }
