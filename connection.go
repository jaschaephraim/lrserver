package lrserver

import (
	"log"

	"code.google.com/p/go.net/websocket"
)

type connection struct {
	websocket  *websocket.Conn
	logger     *log.Logger
	handshake  bool
	helloChan  chan *clientHello
	reloadChan chan string
	alertChan  chan string
	closeChan  chan struct{}
}

func newConnection(ws *websocket.Conn, logger *log.Logger) *connection {
	return &connection{
		websocket: ws,
		logger:    logger,

		helloChan:  make(chan *clientHello),
		reloadChan: make(chan string),
		alertChan:  make(chan string),
		closeChan:  make(chan struct{}),
	}
}

func (c *connection) start() {
	go c.listen()
	go c.respond()
	<-c.closeChan
}

func (c *connection) listen() {
	for {
		hello := new(clientHello)
		err := websocket.JSON.Receive(c.websocket, hello)
		if err != nil {
			c.close()
		}
		c.helloChan <- hello
		break
	}
}

func (c *connection) respond() {
	for {
		var resp interface{}

		select {
		case hello := <-c.helloChan:
			if !validateHello(hello) {
				c.logger.Println("invalid handshake, disconnecting")
				c.close()
				return
			}
			resp = serverHello
			c.handshake = true
		case file := <-c.reloadChan:
			if !c.handshake {
				c.close()
				return
			}
			resp = newServerReload(file)
		case msg := <-c.alertChan:
			if !c.handshake {
				c.close()
				return
			}
			resp = newServerAlert(msg)
		}

		err := websocket.JSON.Send(c.websocket, resp)
		if err != nil {
			c.close()
			return
		}
	}
}

func (c *connection) close() {
	c.closeChan <- struct{}{}
}
