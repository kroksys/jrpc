package jrpc

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/kroksys/jrpc/registry"
	"github.com/kroksys/jrpc/spec"
)

// Server is just a parent for json-rpc server using websockets
type Server struct {
	*registry.Registry
}

// Creates new server with initialised registry
func NewServer() *Server {
	return &Server{
		Registry: registry.NewRegistry(),
	}
}

// Http server handler to upgrade net.Conn to jrpc Conn and
// forwards connection handling to the connection gorutines.
func (s *Server) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("upgrade error: %s", err)
		return
	}
	defer conn.Close()
	s.defaultConnHandler(newConn(conn))
}

// Main handler for jrpc Conn. It does ping, pong, reading, writing
// and parsing incoming messages as jrpc objects.
// When receives jrpc object it tries to execute a method from registry.
func (s *Server) defaultConnHandler(c *Conn) {
	defer c.close()
	pinger := time.NewTicker(pingPeriod)
	defer pinger.Stop()
	for {
		select {
		case msg := <-c.in:
			data, tp := spec.Parse(msg)
			switch tp {
			case spec.TypeRequest:
				go func() {
					request := data.(spec.Request)
					resp := s.Registry.Call(context.TODO(), request, c.Write)
					responseData, err := json.Marshal(resp)
					if err != nil {
						return
					}
					c.out <- responseData
				}()
			case spec.TypeNotification:
				// notification := data.(spec.Notification)
			}
		case notif := <-c.Write:
			go func() {
				responseData, err := json.Marshal(notif)
				if err != nil {
					return
				}
				c.out <- responseData
			}()
		case msg := <-c.out:
			c.write(msg)
		case <-pinger.C:
			c.ping()
		case <-c.exit:
			return
		}
	}
}
