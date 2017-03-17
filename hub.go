package ws

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Client is interface for the client type
type Client interface {
	Write([]byte)
	Listen()
	Close()
}

// ClientFn is the Client creation function callback
type ClientFn func(*Hub, *websocket.Conn, *http.Request) Client

// Connections is a map of Client interfaces to their corresponding WebSocket connection
type Connections map[Client]*websocket.Conn

// Hub is WebSocket hub instance
type Hub struct {
	upgrader     websocket.Upgrader
	once         sync.Once
	clientFn     ClientFn
	writeWait    atomic.Value // time.Duration
	pongWait     atomic.Value // time.Duration
	pingInterval atomic.Value // time.Duration
	readLimit    atomic.Value // int64
	conns        Connections
	ops          chan func(Connections)
	shutdown     chan struct{}
	shutdownInit chan struct{}
}

// New returns a new websocket hub instance.
// NOTE: if fn is nil, the default Client will be used and the
// default read function will Broadcast the message to all
// connected users.
func New(upgrader websocket.Upgrader, fn ClientFn) *Hub {

	h := &Hub{
		upgrader:     upgrader,
		conns:        make(Connections),
		ops:          make(chan func(Connections)),
		shutdownInit: make(chan struct{}),
		shutdown:     make(chan struct{}),
		clientFn:     fn,
	}

	h.SetReadLimit(512)
	h.SetWriteWait(10 * time.Second)
	h.SetPongWait(60 * time.Second)

	if h.clientFn == nil {
		h.clientFn = func(h *Hub, conn *websocket.Conn, r *http.Request) Client {
			fn := func(msg []byte) {
				h.Broadcast(msg)
			}
			return NewClient(h, conn, fn)
		}
	}

	return h
}

// SetPongWait sets the pong wait timeout and automatically sets the
// ping period based off of the duration provided. The WebSocket's
// ReadDeadline will also be set to the same value.
func (h *Hub) SetPongWait(d time.Duration) {
	h.pongWait.Store(d)
	h.pingInterval.Store((d * 9) / 10)
}

// SetWriteWait sets the wait timeout for WebSocket writes.
func (h *Hub) SetWriteWait(d time.Duration) {
	h.writeWait.Store(d)
}

// SetReadLimit sets the limit, in bytes, for read operations
func (h *Hub) SetReadLimit(size int64) {
	h.readLimit.Store(size)
}

// WriteDeadline returns the deadline for connection writes
func (h *Hub) WriteDeadline() time.Time {
	return time.Now().Add(h.writeWait.Load().(time.Duration))
}

// ReadDeadline returns the deadline for connection reads
func (h *Hub) ReadDeadline() time.Time {
	return time.Now().Add(h.pongWait.Load().(time.Duration))
}

// ReadLimit returns the read message limit
func (h *Hub) ReadLimit() int64 {
	return h.readLimit.Load().(int64)
}

// PingInterval returns the ping interval for WebSocket
// ping-pong keepalives.
func (h *Hub) PingInterval() time.Duration {
	return h.pingInterval.Load().(time.Duration)
}

// Accept starts the WebSocket connection and adds connection
// to the pool of clients.
func (h *Hub) Accept(w http.ResponseWriter, r *http.Request) error {

	select {
	case <-h.shutdownInit:
		return errors.New("Hub closed")
	default:
		h.once.Do(func() {
			go h.listen()
		})

		conn, err := h.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return err
		}

		h.add(conn, r)

		return nil
	}
}

// Broadcast sends a message to all currently registered connections
func (h *Hub) Broadcast(msg []byte) {
	h.Do(func(conns Connections) {
		for c := range conns {
			c.Write(msg)
		}
	})
}

// BroadcastTo sends a message to the provided clients only
func (h *Hub) BroadcastTo(msg []byte, filter func(Client) bool) {
	h.Do(func(conns Connections) {
		for c := range conns {
			if filter(c) {
				c.Write(msg)
			}
		}
	})
}

// Shutdown finishes all actions queued prior to calling Shutdown,
// closes all client connections and then stops the Hub
func (h *Hub) Shutdown() {
	h.Do(func(conns Connections) {

		close(h.shutdownInit)

		defer close(h.shutdown)

		for c := range conns {
			delete(conns, c)
			c.Close()
		}
	})

	<-h.shutdown
}

func (h *Hub) add(conn *websocket.Conn, r *http.Request) {
	h.Do(func(conns Connections) {
		c := h.clientFn(h, conn, r)
		conns[c] = conn
		go c.Listen()
	})
}

// Remove removes the provided client from the Hub
// NOTE: this does not close the connection on your behalf
// but rather just removes if from being tracked by the Hub,
// it is your own responsibility to close your own client, and
// is usually done within the Client's Close() function transparently.
func (h *Hub) Remove(c Client) {
	h.Do(func(conns Connections) {
		delete(conns, c)
	})
}

// Do executes the provided function in sequence and is thread safe.
// It allows external logic to be executed from within Hub.
func (h *Hub) Do(fn func(conns Connections)) {
	select {
	case <-h.shutdownInit:
	default:
		h.ops <- fn
	}
}

func (h *Hub) listen() {

	defer close(h.ops)

	var op func(conns Connections)

FOR:
	for {
		select {
		case <-h.shutdown:
			break FOR
		case op = <-h.ops:
			op(h.conns)
		}
	}

	// drain any that may have been queued while closing/shutting down
	// to avoid sending on a closed channel
DRAIN:
	for {
		select {
		case <-h.ops:
		default:
			break DRAIN
		}
	}

	// just in case cleanup
	for c := range h.conns {
		delete(h.conns, c)
	}
}
