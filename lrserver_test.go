package lrserver_test

import (
	"errors"
	"github.com/jaschaephraim/lrserver"
	"golang.org/x/net/websocket"
	"net/http"
	"reflect"
	"testing"
	"time"
)

var clientHello = struct {
	Command   string   `json:"command"`
	Protocols []string `json:"protocols"`
}{
	"hello",
	[]string{
		"http://livereload.com/protocols/official-7",
		"http://livereload.com/protocols/official-8",
		"http://livereload.com/protocols/2.x-origin-version-negotiation",
	},
}

type serverHello struct {
	Command    string   `json:"command"`
	Protocols  []string `json:"protocols"`
	ServerName string   `json:"serverName"`
}

type serverReload struct {
	Command string `json:"command"`
	Path    string `json:"path"`
	LiveCSS bool   `json:"liveCSS"`
}

type serverAlert struct {
	Command string `json:"command"`
	Message string `json:"message"`
}

func TestDisconnectedClose(t *testing.T) {
	lrserver.Close()
}

func TestDisconnectedReload(t *testing.T) {
	lrserver.Reload("")
}

func TestDisconnectedAlert(t *testing.T) {
	lrserver.Alert("")
}

func TestListenAndServe(t *testing.T) {
	connect(t)
	close(t)
}

func TestClose(t *testing.T) {
	connect(t)
	close(t)
	_, err := dial()
	if err == nil {
		t.Fatal("unsuccessful closing of server")
	}
}

func TestJS(t *testing.T) {
	start(t)
	resp, err := http.Get("http://localhost:35729/livereload.js")
	if err != nil {
		t.Fatal(err)
	}
	close(t)
	bytes := make([]byte, 65536)
	i, _ := resp.Body.Read(bytes)
	js := string(bytes[:i])

	if js != lrserver.JS {
		t.Fatal("unsuccessful serving of javascript")
	}
}

func TestHandshake(t *testing.T) {
	ws := connect(t)
	err := handshake(ws, t)
	if err != nil {
		t.Fatal(err)
	}
	close(t)
}

func TestReload(t *testing.T) {
	ws := connect(t)
	err := handshake(ws, t)
	if err != nil {
		t.Fatal(err)
	}

	lrserver.Reload("index.html")
	sr := new(serverReload)
	websocket.JSON.Receive(ws, sr)
	close(t)

	if !reflect.DeepEqual(*sr, serverReload{
		"reload",
		"index.html",
		true,
	}) {
		t.Fatal("unsuccessful reload")
	}
}

func TestAlert(t *testing.T) {
	ws := connect(t)
	err := handshake(ws, t)
	if err != nil {
		t.Fatal(err)
	}

	lrserver.Alert("danger danger")
	sa := new(serverAlert)
	websocket.JSON.Receive(ws, sa)
	close(t)

	if !reflect.DeepEqual(*sa, serverAlert{
		"alert",
		"danger danger",
	}) {
		t.Fatal("unsuccessful alert")
	}
}

func TestReject(t *testing.T) {
	ws := connect(t)
	websocket.JSON.Send(ws, struct{ string }{"bingo"})
	err := handshake(ws, t)
	if err == nil {
		t.Fatal("unsuccessful reject")
	}
	close(t)
}

func connect(t *testing.T) *websocket.Conn {
	start(t)
	ws, err := dial()
	for i := 0; i < 3 && err != nil; i++ {
		time.Sleep(time.Millisecond * 500)
		ws, err = dial()
	}
	if err != nil {
		t.Fatal(err)
	}
	return ws
}

func start(t *testing.T) {
	go func() {
		err := lrserver.ListenAndServe()
		if err != nil {
			t.Fatal(err)
		}
	}()
}

func close(t *testing.T) {
	err := lrserver.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func dial() (*websocket.Conn, error) {
	return websocket.Dial("ws://localhost:35729/livereload", "", "http://localhost/")
}

func handshake(ws *websocket.Conn, t *testing.T) error {
	websocket.JSON.Send(ws, clientHello)
	sh := new(serverHello)
	websocket.JSON.Receive(ws, sh)
	if !reflect.DeepEqual(*sh, serverHello{
		"hello",
		[]string{
			"http://livereload.com/protocols/official-7",
			"http://livereload.com/protocols/official-8",
			"http://livereload.com/protocols/official-9",
			"http://livereload.com/protocols/2.x-origin-version-negotiation",
			"http://livereload.com/protocols/2.x-remote-control",
		},
		"collective-dev",
	}) {
		return errors.New("unsuccessful handshake")
	}
	return nil
}
