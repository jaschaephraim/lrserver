# `lrserver` LiveReload server for Go #

Golang package that implements a simple LiveReload server as described in the [LiveReload protocol](http://feedback.livereload.com/knowledgebase/articles/86174-livereload-protocol).

```bash
go get github.com/jaschaephraim/lrserver
```

Using the default address of ":35729" (which can be changed by setting `lrserver.Addr`):

- `http://localhost:35729/livereload.js` serves the LiveReload client JavaScript (https://github.com/livereload/livereload-js, which can be changed by setting `lrserver.JS`),

- `ws://localhost:35729/livereload` communicates with the client.

File watching must be implemented by your own application, and reload/alert
requests sent programmatically by calling `lrserver.Reload(file string)` and
`lrserver.Alert(msg string)`.

## Usage [![GoDoc](https://godoc.org/github.com/jaschaephraim/lrserver?status.svg)](http://godoc.org/github.com/jaschaephraim/lrserver) ##

### Functions ###

```go
func ListenAndServe() error
```

ListenAndServe starts the server at `lrserver.Addr`.

```go
func Close() error
```

Close ungracefully stops the currently running server.

```go
func Reload(file string)
```

Reload sends a reload request to connected client.

```go
func Alert(msg string)
```

Alert sends an alert request to connected client.

### Variables ###

```go
var (
    // Addr is typically just the port number where the LiveReload server can be reached.
    Addr = ":35729"

    // LiveCSS tells LiveReload whether you want it to update CSS without reloading
    LiveCSS = true

    // JS is initialized to contain LiveReload's client JavaScript
    // (https://github.com/livereload/livereload-js)
    JS string
)
```

## Example ##

```go
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
```
