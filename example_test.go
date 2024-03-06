package lrserver_test

import (
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/jaschaephraim/lrserver"
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
	err = watcher.Add("/path/to/watched/dir")
	if err != nil {
		log.Fatalln(err)
	}

	// Start LiveReload server
	lr := lrserver.New(lrserver.DefaultName, lrserver.DefaultPort)
	go lr.ListenAndServe()

	// Start goroutine that requests reload upon watcher event
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				lr.Reload(event.Name)
			case err := <-watcher.Errors:
				log.Println(err)
			}
		}
	}()

	// Start serving html
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(html))
	})
	http.ListenAndServe(":3000", nil)
}
