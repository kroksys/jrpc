package jrpc

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/kroksys/jrpc/conn"
	"github.com/kroksys/jrpc/registry"
	"github.com/kroksys/jrpc/spec"
)

const (
	pingPeriod = time.Second * 30
)

// Server is just a parent for json-rpc server using websockets
type Server struct {
	*registry.Registry
	LogsOn bool
}

// Creates new server with initialised registry
func NewServer(logsOn bool) *Server {
	return &Server{
		Registry: registry.NewRegistry(logsOn),
		LogsOn:   logsOn,
	}
}

// Main handler for jrpc Conn. It does ping, pong, reading, writing
// and parsing incoming messages as jrpc objects.
// When receives jrpc object it tries to execute a method from registry.
func (s *Server) defaultConnHandler(c *conn.Conn, ctx context.Context) {
	defer c.Close()
	pinger := time.NewTicker(pingPeriod)
	defer pinger.Stop()
	for {
		select {
		case <-pinger.C:
			c.Ping()
		case _, running := <-c.Exit:
			if !running {
				return
			}
		case msg := <-c.In:
			go func() {
				data, tp := spec.Parse(msg)
				switch tp {
				case spec.TypeRequest:
					request := data.(spec.Request)
					if s.LogsOn {
						log.Printf("Request Id:%v Method:%s Params: %v\n", request.ID, request.Method, request.Params)
					}
					resp, shouldReply := s.Registry.Call(ctx, request, c)
					if shouldReply {
						responseData, err := json.Marshal(resp)
						if err != nil {
							return
						}
						if s.LogsOn {
							log.Printf("Response Id:%v Result:%v Err: %v\n", resp.ID, resp.Result, resp.Error)
						}
						c.Send(responseData)
					}
				case spec.TypeNotification:
					notification := data.(spec.Notification)
					if s.LogsOn {
						log.Printf("Method:%s Params: %v\n", notification.Method, notification.Params)
					}
					err := s.Registry.Subscribe(ctx, notification, c)
					if err != nil {
						if s.LogsOn {
							log.Printf("%s:error: %v\n", notification.Method, err)
						}
						errData, err := json.Marshal(err)
						if err != nil {
							if s.LogsOn {
								log.Printf("%s:json.Marshal error: %s\n", notification.Method, err.Error())
							}
							return
						}
						c.Send(errData)
					}
				}
			}()
		}
	}
}

// Go gin handler. There is a bug that this handler does not work
// with gin Group. Have no idea why. So its mandatory to use
// gin router.GET() to register the route.
func (s *Server) WebsocketHandlerGin(g *gin.Context) {
	cn, _, _, err := ws.UpgradeHTTP(g.Request, g.Writer)
	if err != nil {
		log.Printf("upgrade error: %s", err)
		return
	}
	defer cn.Close()
	s.defaultConnHandler(conn.NewConn(cn), g)
}

// Http server handler to upgrade net.Conn to jrpc Conn and
// forwards connection handling to the connection gorutines.
func (s *Server) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	cn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("upgrade error: %s", err)
		return
	}
	defer cn.Close()
	s.defaultConnHandler(conn.NewConn(cn), r.Context())
}
