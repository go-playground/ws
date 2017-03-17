package ws

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	empty = []byte{}
)

type client struct {
	hub    *Hub
	conn   *websocket.Conn
	readFn ReadFn
	send   chan []byte
	close  chan struct{}
	once   sync.Once
}

// ReadFn is the function to be called directly after a read operation
// by the default client.
type ReadFn func([]byte)

// NewClient returns a new instance of the default client.
func NewClient(h *Hub, conn *websocket.Conn, readFn ReadFn) Client {
	return &client{
		hub:    h,
		conn:   conn,
		readFn: readFn,
		close:  make(chan struct{}),
		send:   make(chan []byte),
	}
}

func (c *client) Listen() {
	go c.write()
	c.read()
}

// Close closes the connection
func (c *client) Close() {
	c.once.Do(func() {
		c.conn.Close()
		close(c.close)
		c.hub.Remove(c)
	})
}

func (c *client) read() {

	defer func() {
		c.Close()
	}()

	c.conn.SetReadLimit(c.hub.ReadLimit())

	err := c.conn.SetReadDeadline(c.hub.ReadDeadline())
	if err != nil {
		log.Printf("read deadline reached '%s'\n", err)
		return
	}

	c.conn.SetPongHandler(func(string) error {
		err := c.conn.SetReadDeadline(c.hub.ReadDeadline())
		if err != nil {
			log.Printf("error in pong handler '%s'\n", err)
		}
		return nil
	})

	for {

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v\n", err)
			}
			break
		}

		c.readFn(message)
	}
}

func (c *client) write() {

	ticker := time.NewTicker(c.hub.PingInterval())
	defer func() {
		c.Close()
		ticker.Stop()
		close(c.send)
	}()

FOR:
	for {

		select {
		case <-c.close:
			break FOR
		case <-ticker.C:

			err := c.conn.SetWriteDeadline(c.hub.WriteDeadline())
			if err != nil {
				log.Printf("error setting write deadline '%s'\n", err)
				break FOR
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, empty); err != nil {
				log.Printf("error sending ping message '%s'\n", err)
				break FOR
			}

		case msg := <-c.send:

			err := c.conn.SetWriteDeadline(c.hub.WriteDeadline())
			if err != nil {
				log.Printf("error setting write deadline '%s'\n", err)
				break FOR
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("error retrieving next writer '%s'\n", err)
				break FOR
			}

			_, err = w.Write(msg)
			if err != nil {
				log.Printf("error writing message '%s'\n", err)
			}

			if err = w.Close(); err != nil {
				break FOR
			}
		}
	}
}

func (c *client) Write(msg []byte) {
	select {
	case <-c.close:
	default:
		c.send <- msg
	}
}
