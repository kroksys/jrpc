package jrpc

import (
	"log"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/kroksys/jrpc/registry"
)

type Server struct {
	*registry.Registry
}

func NewServer() *Server {
	return &Server{
		Registry: registry.NewRegistry(),
	}
}

func (s *Server) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("upgrade error: %s", err)
		return
	}
	defer conn.Close()
	c := newConn(conn, s.Registry)
	c.defaultHandler()
}
