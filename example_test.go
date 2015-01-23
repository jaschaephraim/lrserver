package lrserver_test

import (
	"log"
	"net/http"

	"github.com/jaschaephraim/lrserver"
	"golang.org/x/exp/fsnotify"
)

// html includes the client JavaScript
const html = `<!doctype html>
<html>
<head>
  <title>Example</title>
</head>
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

	// Watch dir
	err = watcher.Watch("/path/to/watched/dir")
	if err != nil {
		log.Fatalln(err)
	}

	// Start LiveReload server
	lr, err := lrserver.New(lrserver.DefaultName, lrserver.DefaultPort)
	if err != nil {
		log.Fatalln(err)
	}
	go lr.ListenAndServe()

	// Start goroutine that requests reload upon watcher event
	go func() {
		for {
			event := <-watcher.Event
			lr.Reload(event.Name)
		}
	}()

	// Start serving html
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(html))
	})
	http.ListenAndServe(":3000", nil)
}
