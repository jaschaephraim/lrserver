package lrserver

import (
	"errors"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
)

type server struct {
	server     *http.Server
	listener   *net.Listener
	connection *connection

	closing bool
}

func newServer() *server {
	s := server{
		server: &http.Server{
			Addr: Addr,
		},
	}
	return &s
}

func (s *server) listenAndServe() error {
	// Create listener
	l, err := net.Listen("tcp", Addr)
	if err != nil {
		return err
	}
	s.listener = &l

	// Start server
	err = s.server.Serve(*s.listener)
	if err != nil && !s.closing {
		return err
	}

	s.closing = false
	return nil
}

func (s *server) close() error {
	if s.listener == nil {
		return errors.New("close called before server started")
	}

	s.closing = true
	err := (*s.listener).Close()
	if err != nil {
		return err
	}

	s.listener = nil
	return nil
}

func (s *server) setConnection(ws *websocket.Conn) {
	s.connection = newConnection(ws)
	s.connection.start()
}

func (s *server) sendReload(file string) {
	if s.connection == nil {
		logger.Printf("can't send request to reload %s, no connection", file)
		return
	}
	s.connection.reloadChan <- file
}

func (s *server) sendAlert(msg string) {
	if s.connection == nil {
		logger.Printf("can't send request to alert \"%s\", no connection", msg)
		return
	}
	s.connection.alertChan <- msg
}
