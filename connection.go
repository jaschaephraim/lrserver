package lrserver

import "code.google.com/p/go.net/websocket"

type connection struct {
	websocket  *websocket.Conn
	handshake  bool
	helloChan  *chan *clientHello
	reloadChan *chan string
	alertChan  *chan string
	closeChan  *chan struct{}
}

func newConnection(ws *websocket.Conn) *connection {
	c := connection{
		websocket: ws,
	}

	hc := make(chan *clientHello)
	rc := make(chan string)
	ac := make(chan string)
	cc := make(chan struct{})
	c.helloChan = &hc
	c.reloadChan = &rc
	c.alertChan = &ac
	c.closeChan = &cc

	return &c
}

func (c *connection) start() {
	go c.listen()
	go c.respond()
	<-*c.closeChan
}

func (c *connection) listen() {
	for {
		hello := new(clientHello)
		err := websocket.JSON.Receive(c.websocket, hello)
		if err != nil {
			c.close()
		}
		*c.helloChan <- hello
		break
	}
}

func (c *connection) respond() {
	for {
		var resp interface{}

		select {
		case hello := <-*c.helloChan:
			if !validateHello(hello) {
				logger.Println("invalid handshake, disconnecting")
				c.close()
				return
			}
			resp = serverHello
			c.handshake = true
		case file := <-*c.reloadChan:
			if !c.handshake {
				c.close()
				return
			}
			resp = newServerReload(file)
		case msg := <-*c.alertChan:
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
	*c.closeChan <- struct{}{}
}
