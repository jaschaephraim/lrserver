/*
Package lrserver implements a basic LiveReload server.

(See http://feedback.livereload.com/knowledgebase/articles/86174-livereload-protocol .)

Using the default address of ":35729" (which can be changed by setting lrserver.Addr):

  http://localhost:35729/livereload.js

serves the LiveReload client JavaScript (https://github.com/livereload/livereload-js,
which can be changed by setting lrserver.JS),

  ws://localhost:35729/livereload

communicates with the client.

File watching must be implemented by your own application, and reload/alert
requests sent programmatically by calling lrserver.Reload(file string) and
lrserver.Alert(msg string).
*/
package lrserver

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"code.google.com/p/go.net/websocket"
)

var (
	// Addr is typically just the port number where the LiveReload server can be reached.
	Addr = ":35729"

	// LiveCSS tells LiveReload whether you want it to update CSS without reloading
	LiveCSS = true
)

type LRServer struct {
	httpServer *http.Server
	listener   net.Listener
	connection *connection

	logger *log.Logger

	Running chan struct{}
}

func NewLRServer(writer io.Writer) (srv *LRServer, err error) {
	srv = &LRServer{
		httpServer: &http.Server{
			Addr: Addr,
		},
	}

	if writer == nil {
		writer = os.Stdout
	}

	srv.logger = log.New(writer, "[lrserver] ", 0)

	r := mux.NewRouter()

	// Handle JS request
	r.HandleFunc("/livereload.js", srv.jsHandler)

	// Handle WebSockets
	r.Handle("/livereload", websocket.Handler(func(ws *websocket.Conn) {
		srv.connection = newConnection(ws, srv.logger)
		srv.connection.start()
	}))

	srv.logger.Println("listening on " + Addr)

	errc := make(chan error)

	go func() {

		// Create listener
		srv.listener, err = net.Listen("tcp", Addr)
		if err != nil {
			errc <- err
			return
		}

		// Start server
		if err = http.Serve(srv.listener, r); err != nil {
			errc <- err
		}
	}()

	select {
	case e := <-errc:
		if e != nil {
			return nil, e
		}
	case <-time.After(time.Millisecond * 150):
		srv.logger.Println("No error, continueing")
		srv.Running = make(chan struct{})
	}

	return srv, nil
}

func (s *LRServer) jsHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/javascript")

	_, err := rw.Write([]byte(JS))
	if err != nil {
		s.logger.Println(err)
	}
}

// Close ungracefully stops the currently running server.
func (s *LRServer) Close() error {
	s.logger.Println("stopping server")

	if err := s.listener.Close(); err != nil {
		return err
	}

	close(s.Running)
	return nil
}

// Reload sends a reload request to connected client.
func (s *LRServer) Reload(file string) {
	if s.connection == nil {
		s.logger.Println("Connection not ready yet.")
		return
	}
	s.logger.Println("requesting reload: " + file)
	s.connection.reloadChan <- file
}

// Alert sends an alert request to connected client.
func (s *LRServer) Alert(msg string) {
	if s.connection == nil {
		s.logger.Println("Connection not ready yet.")
		return
	}
	s.logger.Println("requesting alert: " + msg)
	s.connection.alertChan <- msg
}
