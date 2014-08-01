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
	"log"
	"net/http"
	"os"

	"code.google.com/p/go.net/websocket"
)

var (
	// Addr is typically just the port number where the LiveReload server can be reached.
	Addr = ":35729"

	// LiveCSS tells LiveReload whether you want it to update CSS without reloading
	LiveCSS = true

	// JS is initialized to contain LiveReload's client JavaScript (https://github.com/livereload/livereload-js)
	JS string

	logger = log.New(os.Stdout, "[lrserver] ", 0)
	srv    = newServer()
)

func init() {
	// Handle JS request
	http.HandleFunc("/livereload.js", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/javascript")
		_, err := rw.Write([]byte(JS))
		if err != nil {
			logger.Println(err)
		}
	})

	// Handle WebSockets
	http.Handle("/livereload", websocket.Handler(func(ws *websocket.Conn) {
		srv.setConnection(ws)
	}))
}

// ListenAndServe starts the server at lrserver.Addr.
func ListenAndServe() error {
	logger.Println("listening on " + Addr)
	return srv.listenAndServe()
}

// Close ungracefully stops the currently running server.
func Close() error {
	logger.Println("stopping server")
	return srv.close()
}

// Reload sends a reload request to connected client.
func Reload(file string) {
	logger.Println("requesting reload: " + file)
	srv.sendReload(file)
}

// Alert sends an alert request to connected client.
func Alert(msg string) {
	logger.Println("requesting alert: " + msg)
	srv.sendAlert(msg)
}
