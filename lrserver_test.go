package lrserver_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jaschaephraim/lrserver"
	. "github.com/smartystreets/goconvey/convey"
)

const localhost string = "://127.0.0.1"

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

var randomMessage = struct {
	Command string `json:"command"`
}{
	"random",
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

func Test(t *testing.T) {
	Convey("Given a new server", t, func() {
		srv := lrserver.New(lrserver.DefaultName, 0)

		Convey("StatusLog() and ErrorLog() should return loggers", func() {
			logger := log.New(nil, "", 0)
			So(srv.StatusLog(), ShouldHaveSameTypeAs, logger)
			So(srv.ErrorLog(), ShouldHaveSameTypeAs, logger)
		})

		srv.SetStatusLog(nil)
		srv.SetErrorLog(nil)

		// Start server
		Convey("that is running", func() {
			go srv.ListenAndServe()

			time.Sleep(time.Millisecond)

			Convey("a dynamically assigned port should be updated", func() {
				So(srv.Port(), ShouldNotEqual, 0)
			})

			// Test JS
			Convey("JS should be served successfully", func() {
				client := new(http.Client)
				resp, err := client.Get(
					fmt.Sprintf("http%s:%d/livereload.js", localhost, srv.Port()),
				)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}

				bodyString := string(body)
				So(bodyString, ShouldEqual, lrserver.JS)
			})

			// Connect WebSocket
			Convey("and a connected websocket", func() {
				dialer := new(websocket.Dialer)
				conn, resp, err := dialer.Dial(
					fmt.Sprintf("ws%s:%d/livereload", localhost, srv.Port()),
					http.Header{},
				)
				if err != nil {
					t.Fatal(err)
				}

				So(resp.StatusCode, ShouldEqual, 101)

				// Receive hello
				hello := new(serverHello)
				err = conn.ReadJSON(hello)
				if err != nil {
					t.Fatal(err)
				}

				So(*hello, ShouldResemble, serverHello{
					"hello",
					[]string{
						"http://livereload.com/protocols/official-7",
						"http://livereload.com/protocols/official-8",
						"http://livereload.com/protocols/official-9",
						"http://livereload.com/protocols/2.x-origin-version-negotiation",
						"http://livereload.com/protocols/2.x-remote-control",
					},
					srv.Name(),
				})

				Convey("an unconnected server should be detected", func() {
					isConnected := srv.IsConnected()
					So(isConnected, ShouldBeFalse)
				})

				// Test bad handshake
				Convey("an invalid handshake should close the connection", func() {
					err = conn.WriteJSON(randomMessage)
					if err != nil {
						t.Fatal(err)
					}

					_, _, err := conn.NextReader()
					So(reflect.TypeOf(err).String(), ShouldEqual, "*websocket.CloseError")
				})

				// Send valid handshake
				Convey("and a successful handshake", func() {
					err = conn.WriteJSON(clientHello)
					if err != nil {
						t.Fatal(err)
					}

					time.Sleep(time.Millisecond)

					Convey("a connected server should be detected", func() {
						isConnected := srv.IsConnected()
						So(isConnected, ShouldBeTrue)
					})

					Convey("a valid client message should be tolerated", func() {
						err = conn.WriteJSON(randomMessage)
						if err != nil {
							t.Fatal(err)
						}

						errChan := make(chan error)
						failed := false
						go func() {
							_, _, err := conn.NextReader()
							errChan <- err
						}()
						go func() {
							<-errChan
							failed = true
						}()

						time.Sleep(time.Millisecond)

						So(failed, ShouldBeFalse)
					})

					// Test reload
					Convey("reload should work", func() {
						file := "file"
						srv.Reload(file)

						sr := new(serverReload)
						err = conn.ReadJSON(sr)
						if err != nil {
							t.Fatal(err)
						}

						So(*sr, ShouldResemble, serverReload{
							"reload",
							file,
							srv.LiveCSS(),
						})
					})

					// Test alert
					Convey("alert should work", func() {
						msg := "alert"
						srv.Alert(msg)

						sa := new(serverAlert)
						err = conn.ReadJSON(sa)
						if err != nil {
							t.Fatal(err)
						}

						So(*sa, ShouldResemble, serverAlert{
							"alert",
							msg,
						})
					})

					// Close
					Convey("close should work", func() {
						err = srv.Close()
						So(err, ShouldEqual, nil)

						client := new(http.Client)
						_, err := client.Get(
							fmt.Sprintf("http%s:%d/livereload.js", localhost, srv.Port()),
						)
						So(err, ShouldNotEqual, nil)
					})

					Convey("Create another new server", func() {
						srv := lrserver.New(lrserver.DefaultName, 0)

						Convey("StatusLog() and ErrorLog() should return loggers", func() {
							logger := log.New(nil, "", 0)
							So(srv.StatusLog(), ShouldHaveSameTypeAs, logger)
							So(srv.ErrorLog(), ShouldHaveSameTypeAs, logger)
						})

						srv.SetStatusLog(nil)
						srv.SetErrorLog(nil)

						// Start server
						Convey("that is running", func() {
							go srv.ListenAndServe()

							time.Sleep(time.Millisecond)

							Convey("a dynamically assigned port should be updated", func() {
								So(srv.Port(), ShouldNotEqual, 0)
							})

							// Test JS
							Convey("JS should be served successfully", func() {
								client := new(http.Client)
								resp, err := client.Get(
									fmt.Sprintf("http%s:%d/livereload.js", localhost, srv.Port()),
								)
								if err != nil {
									t.Fatal(err)
								}
								defer resp.Body.Close()

								body, err := io.ReadAll(resp.Body)
								if err != nil {
									t.Fatal(err)
								}

								bodyString := string(body)
								So(bodyString, ShouldEqual, lrserver.JS)
							})

							// Shutdown
							Convey("Shutdown should work", func() {
								err = srv.Shutdown(context.Background())
								So(err, ShouldEqual, nil)

								client := new(http.Client)
								_, err := client.Get(
									fmt.Sprintf("http%s:%d/livereload.js", localhost, srv.Port()),
								)
								So(err, ShouldNotEqual, nil)
							})
						})
					})
				})
			})
		})
	})
}
