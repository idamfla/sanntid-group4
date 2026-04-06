package server

func (srv *Server) ResetState() {
	srv.state.Reset()
}

func (srv *Server) IsMaster() bool {
	return srv.state.GetIsMaster()
}

func (srv *Server) setSelfAsMaster() {
	srv.state.SetMaster()
}

func (srv *Server) setMasterSearch() {
	srv.state.SetMasterSearch()
}

func (srv *Server) isSearchingForMaster() bool { // TODO this could probably be part of the masterElectSession
	return srv.state.IsSearchingForMaster()
}

func (srv *Server) clearMasterSearch() {
	srv.state.ClearMasterSearch()
}

func (srv *Server) isSynced() bool {
	return srv.state.GetSynced()
}

func (srv *Server) setSynced()   { srv.state.SetSynced() }
func (srv *Server) clearSynced() { srv.state.ClearSynced() }
