package server

import "math"

const (
	quorumPercent = float64(0.5)
)

func (srv *Server) getQuorum() int {
	return int(math.Ceil(float64(srv.activePeerCount()) * quorumPercent))
}
