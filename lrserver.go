/*
Package lrserver implements a basic LiveReload server.

(See http://feedback.livereload.com/knowledgebase/articles/86174-livereload-protocol .)

Using the default port of ":35729":

  http://localhost:35729/livereload.js

serves the LiveReload client JavaScript, and:

  ws://localhost:35729/livereload

communicates with the client.

File watching must be implemented by your own application, and reload/alert
requests sent programmatically.

Multiple servers can be instantiated, and each can support multiple connections.
*/
package lrserver

const (
	DefaultName string = "LiveReload"
	DefaultPort uint16 = 35729
)
