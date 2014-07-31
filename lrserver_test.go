package lrserver_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"code.google.com/p/go.net/websocket"
	"github.com/jaschaephraim/lrserver"
	. "github.com/smartystreets/goconvey/convey"
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

func TestAll(t *testing.T) {
	Convey("Start a NewServer()", t, func() {
		var logBuf bytes.Buffer
		lrs, err := lrserver.NewLRServer(&logBuf)
		So(err, ShouldBeNil)
		So(lrs, ShouldNotBeNil)

		Convey("create a client", func() {
			ws, err := websocket.Dial("ws://localhost:35729/livereload", "", "http://localhost/")
			So(err, ShouldBeNil)

			Convey("Reject Handshake", func() {
				err := websocket.JSON.Send(ws, struct{ string }{"bingo"})
				So(err, ShouldBeNil)

				sh := new(serverHello)
				err = websocket.JSON.Receive(ws, sh)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "EOF")
			})

			Convey("Handhsake", func() {
				err := websocket.JSON.Send(ws, clientHello)
				So(err, ShouldBeNil)

				sh := new(serverHello)
				err = websocket.JSON.Receive(ws, sh)
				So(err, ShouldBeNil)

				So(sh, ShouldResemble, &serverHello{
					"hello",
					[]string{
						"http://livereload.com/protocols/official-7",
						"http://livereload.com/protocols/official-8",
						"http://livereload.com/protocols/official-9",
						"http://livereload.com/protocols/2.x-origin-version-negotiation",
						"http://livereload.com/protocols/2.x-remote-control",
					},
					"collective-dev",
				})

				Convey("Reload", func() {
					fname := "index.html"
					lrs.Reload(fname)

					sr := new(serverReload)
					err := websocket.JSON.Receive(ws, sr)
					So(err, ShouldBeNil)

					So(sr, ShouldResemble, &serverReload{
						"reload",
						fname,
						true,
					})
				})

				Convey("Alert", func() {
					altext := "danger danger"
					lrs.Alert(altext)

					sa := new(serverAlert)
					err := websocket.JSON.Receive(ws, sa)
					So(err, ShouldBeNil)

					So(sa, ShouldResemble, &serverAlert{
						"alert",
						altext,
					})
				})
			})
		})

		Convey("JS", func() {
			resp, err := http.Get("http://localhost:35729/livereload.js")
			So(err, ShouldBeNil)

			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			jsBody, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(string(jsBody), ShouldEqual, lrserver.JS)
		})

		// test close
		Reset(func() {
			So(logBuf.String(), ShouldStartWith, "[lrserver] listening on :35729\n[lrserver] No error, continueing\n")
			So(lrs.Close(), ShouldBeNil)
		})
	})
}
