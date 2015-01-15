package lrserver_test

import (
	"github.com/jaschaephraim/lrserver"
	"golang.org/x/exp/fsnotify"
	"log"
	"net/http"
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

	// Start LiveReload server
	go lrserver.ListenAndServe()

	// Start goroutine that requests reload upon watcher event
	go func() {
		for {
			event := <-watcher.Events
			lrserver.Reload(event.Name)
		}
	}()

	// Start serving html
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(html))
	})
	http.ListenAndServe(":3000", nil)
}
