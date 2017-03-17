package ws

import (
	"log"
	"sync"
	"testing"

	"net/http"
	"net/http/httptest"

	"strings"

	"github.com/gorilla/websocket"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html

func TestSingleConnection(t *testing.T) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	hub := New(upgrader, nil)
	defer hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		err := hub.Accept(w, r)
		if err != nil {
			t.Fatalf("accepting WebSocket connection '%s'\n", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	msg := "Test Message\n"
	url := strings.Replace(server.URL, "http", "ws", 1) + "/ws"

	var dialer *websocket.Dialer

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {

		defer wg.Done()

		err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			t.Fatal(err)
		}
	}()

	wg.Wait()

	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	resp := string(message)

	if msg != resp {
		log.Fatalf("Expected '%s' Got '%s'", msg, resp)
	}

	// explicitly calling before conn.Close() otherwise an
	// unexpected error will print from closing conn
	hub.Shutdown()
}

func TestMultipleConnections(t *testing.T) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	clientFn := func(h *Hub, conn *websocket.Conn, r *http.Request) Client {

		fn := func(msg []byte) {
			h.BroadcastTo(msg, func(c Client) bool {
				return true
			})
		}

		return NewClient(h, conn, fn)
	}

	hub := New(upgrader, clientFn)
	defer hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		err := hub.Accept(w, r)
		if err != nil {
			t.Fatalf("accepting WebSocket connection '%s'\n", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	msg := "Test Message\n"
	url := strings.Replace(server.URL, "http", "ws", 1) + "/ws"

	var dialer1 *websocket.Dialer
	var dialer2 *websocket.Dialer

	conn1, _, err := dialer1.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn1.Close()

	conn2, _, err := dialer2.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {

		defer wg.Done()

		err = conn1.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			t.Fatal(err)
		}
	}()

	wg.Wait()

	_, message1, err := conn1.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	_, message2, err := conn2.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	resp1 := string(message1)
	resp2 := string(message2)

	if msg != resp1 {
		log.Fatalf("Expected '%s' Got '%s'", msg, resp1)
	}

	if msg != resp2 {
		log.Fatalf("Expected '%s' Got '%s'", msg, resp2)
	}

	// explicitly calling before conn.Close() otherwise an
	// unexpected error will print from closing conn
	hub.Shutdown()
}

func TestCloseAndRemoveConnection(t *testing.T) {

	type myCustomClient struct {
		Client
		id string
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	var c1, c2 Client

	clientFn := func(h *Hub, conn *websocket.Conn, r *http.Request) Client {

		var myClient *myCustomClient

		fn := func(msg []byte) {
			h.BroadcastTo(msg, func(c Client) bool {
				return true
			})
		}

		// create my custom client and add additional information
		// in this case the id of the client as an example...most
		// likely this would be a user ID to identify the user connected.
		myClient = &myCustomClient{
			Client: NewClient(h, conn, fn),
			id:     r.URL.Query().Get("id"),
		}

		if myClient.id == "1" {
			c1 = myClient
		} else {
			c2 = myClient
		}

		return myClient
	}

	hub := New(upgrader, clientFn)
	defer hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		err := hub.Accept(w, r)
		if err != nil {
			t.Fatalf("accepting WebSocket connection '%s'\n", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	msg := "Test Message\n"
	url := strings.Replace(server.URL, "http", "ws", 1) + "/ws?id="

	var dialer1 *websocket.Dialer
	var dialer2 *websocket.Dialer

	conn1, _, err := dialer1.Dial(url+"1", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn1.Close()

	conn2, _, err := dialer2.Dial(url+"2", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {

		defer wg.Done()

		err = conn1.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			t.Fatal(err)
		}
	}()

	wg.Wait()

	_, message1, err := conn1.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	_, message2, err := conn2.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	resp1 := string(message1)
	resp2 := string(message2)

	if msg != resp1 {
		log.Fatalf("Expected '%s' Got '%s'", msg, resp1)
	}

	if msg != resp2 {
		log.Fatalf("Expected '%s' Got '%s'", msg, resp2)
	}

	c1.Close()

	hub.Remove(c2)
	c2.Close()

	// explicitly calling before conn.Close() otherwise an
	// unexpected error will print from closing conn
	hub.Shutdown()
}
