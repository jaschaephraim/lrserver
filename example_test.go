package lrserver_test

import (
	"log"
	"net/http"

	"github.com/jaschaephraim/lrserver"
	"gopkg.in/fsnotify.v0"
)

// html includes the client JavaScript
const html = `<!doctype html>
<html>
<head>
  <title>Example</title>
<body>
  <script src="http://localhost:35729/livereload.js"></script>
</body>
</html>`

func Example() {
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()

	// Add dir to watcher
	err = watcher.Add("/path/to/watched/dir")
	if err != nil {
		log.Fatalln(err)
	}

	lr, err := lrserver.NewLRServer(nil)
	if err != nil {
		log.Fatalln(err)
	}

	// Start goroutine that requests reload upon watcher event
	go func() {
		for {
			event := <-watcher.Events
			lr.Reload(event.Name)
		}
	}()

	// Start serving html
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(html))
	})
	http.ListenAndServe(":3000", nil)

	<-lr.Running
}
