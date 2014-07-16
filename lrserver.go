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
	"code.google.com/p/go.net/websocket"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
)

// JS is initialized to contain LiveReload's client JavaScript (https://github.com/livereload/livereload-js)
var JS string

var (
	Addr    = ":35729"
	LiveCSS = true

	rlFile  string
	rlAlert string

	logger = log.New(os.Stdout, "[lrserver] ", 0)
	srv    = &server{
		mux: http.NewServeMux(),
	}
)

type server struct {
	server   *http.Server
	mux      *http.ServeMux
	listener *net.Listener
	state    http.ConnState

	serve func() error
	close func() error
}

func init() {
	// Init server
	srv.server = &http.Server{
		Handler: srv.mux,
		ConnState: func(c net.Conn, cs http.ConnState) {
			srv.state = cs
		},
	}
	srv.serve = func() error {
		lsn, err := net.Listen("tcp", Addr)
		if err != nil {
			return err
		}
		srv.listener = &lsn
		err = srv.server.Serve(*srv.listener)
		if err != nil && srv.state != http.StateClosed && srv.state != http.StateHijacked {
			return err
		}
		return nil
	}
	srv.close = func() error {
		err := (*srv.listener).Close()
		if err != nil {
			return err
		}
		srv.listener = nil
		return nil
	}

	// JS request handler
	srv.mux.HandleFunc("/livereload.js", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/javascript")
		_, err := rw.Write([]byte(JS))
		if err != nil {
			logger.Println(err)
			return
		}
	})

	// WS handler
	srv.mux.Handle("/livereload", websocket.Handler(func(ws *websocket.Conn) {
		ch := new(clientHello)
		err := websocket.JSON.Receive(ws, ch)
		if err != nil {
			logger.Println(err)
			closeWS(ws)
			return
		}

		if validateClientHello(ch) {
			// Send hello
			err = websocket.JSON.Send(ws, serverHello)
			if err != nil {
				logger.Println(err)
				closeWS(ws)
				return
			}

			// Send reload
			if rlFile != "" {
				err := websocket.JSON.Send(ws, newServerReload(rlFile))
				rlFile = ""
				if err != nil {
					logger.Println(err)
					closeWS(ws)
					return
				}
			}

			// Send alert
			if rlAlert != "" {
				err := websocket.JSON.Send(ws, newServerAlert(rlAlert))
				rlAlert = ""
				if err != nil {
					logger.Println(err)
					closeWS(ws)
					return
				}
			}
		} else {
			err = closeWS(ws)
			if err == nil {
				logger.Println(errors.New("invalid handshake, connection closed"))
			}
			return
		}
	}))
}

// ListenAndServe starts the server at lrserver.Addr.
func ListenAndServe() error {
	logger.Println("listening on " + Addr)
	return srv.serve()
}

// Close ungracefully stops the currently running server.
func Close() error {
	logger.Println("stopping server")
	return srv.close()
}

// Reload sends a reload request to the next incoming WebSocket connection.
func Reload(file string) {
	logger.Println("requesting reload: " + file)
	rlFile = file
}

// Alert sends an alert request to the next incoming WebSocket connection.
func Alert(msg string) {
	logger.Println("requesting alert: " + msg)
	rlAlert = msg
}

var serverHello = struct {
	Command    string   `json:"command"`
	Protocols  []string `json:"protocols"`
	ServerName string   `json:"serverName"`
}{
	"hello",
	[]string{
		"http://livereload.com/protocols/official-7",
		"http://livereload.com/protocols/official-8",
		"http://livereload.com/protocols/official-9",
		"http://livereload.com/protocols/2.x-origin-version-negotiation",
		"http://livereload.com/protocols/2.x-remote-control",
	},
	"collective-dev",
}

type clientHello struct {
	Command   string   `json:"command"`
	Protocols []string `json:"protocols"`
}

type serverReload struct {
	Command string `json:"command"`
	Path    string `json:"path"`
	LiveCSS bool   `json:"liveCSS"`
}

func newServerReload(file string) serverReload {
	return serverReload{
		Command: "reload",
		Path:    file,
		LiveCSS: LiveCSS,
	}
}

type serverAlert struct {
	Command string `json:"command"`
	Message string `json:"message"`
}

func newServerAlert(msg string) serverAlert {
	return serverAlert{
		Command: "alert",
		Message: msg,
	}
}

func validateClientHello(ch *clientHello) bool {
	if ch.Command != "hello" {
		return false
	}
	for _, c := range ch.Protocols {
		for _, s := range serverHello.Protocols {
			if c == s {
				return true
			}
		}
	}
	return false
}

func closeWS(ws *websocket.Conn) error {
	err := ws.Close()
	if err != nil {
		logger.Println(err)
	}
	return err
}
