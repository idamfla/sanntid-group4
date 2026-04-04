package server

func (srv *Server) ResetState() {
	srv.state.Reset()
}

func (srv *Server) IsMaster() bool {
	return srv.state.GetIsMaster()
}

func (srv *Server) SetSelfAsMaster() {
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
	return srv.state.GetIsSynced()
}

func (srv *Server) SetIsSynced(isSynced bool) {
	srv.state.SetIsSynced(isSynced)
}
