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
	s.defaultConnHandler(newConn(conn))
}

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
