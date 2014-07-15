package lrserver_test

import (
	"code.google.com/p/go.net/websocket"
	"errors"
	"github.com/jaschaephraim/lrserver"
	"reflect"
	"testing"
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

func TestHandshake(t *testing.T) {
	ws := connect(t)
	err := handshake(ws, t)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReload(t *testing.T) {
	ws := connect(t)
	lrserver.Reload("index.html")
	err := handshake(ws, t)
	if err != nil {
		t.Fatal(err)
	}

	sr := new(serverReload)
	websocket.JSON.Receive(ws, sr)
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
	lrserver.Alert("danger danger")
	err := handshake(ws, t)
	if err != nil {
		t.Fatal(err)
	}

	sa := new(serverAlert)
	websocket.JSON.Receive(ws, sa)
	if !reflect.DeepEqual(*sa, serverAlert{
		"alert",
		"danger danger",
	}) {
		t.Fatal("unsuccessful alert")
	}
}

func TestReject(t *testing.T) {
	ws := connect(t)
	websocket.JSON.Send(ws, struct{}{})
	err := handshake(ws, t)
	if err == nil {
		t.Fatal("unsuccessful reject")
	}
}

func connect(t *testing.T) *websocket.Conn {
	go lrserver.ListenAndServe()
	ws, err := websocket.Dial("ws://localhost:35729/livereload", "", "http://localhost/")
	if err != nil {
		t.Fatal("unable to establish connection")
	}
	return ws
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
