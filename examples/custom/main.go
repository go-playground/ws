package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/ws"
	"github.com/gorilla/websocket"
)

var (
	hub      *ws.Hub
	tpls     *template.Template
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type myCustomClient struct {
	ws.Client
	id string
}

func main() {

	var err error

	tpls, err = template.New("root").Parse(html)
	if err != nil {
		log.Fatal(err)
	}

	clientFn := func(h *ws.Hub, conn *websocket.Conn, r *http.Request) ws.Client {

		var myClient *myCustomClient

		fn := func(msg []byte) {
			log.Println("Message Recieved:", string(msg))

			// do some custom logic based on the message eg. broadcast to
			// everyone except yourself.
			// This can easily be extended to message a group of users, when
			// a particular even happens server side
			h.BroadcastTo(msg, func(c ws.Client) bool {
				return c.(*myCustomClient).id != myClient.id
			})
		}

		// create my custom client and add additional information
		// in this case the id of the client as an example...most
		// likely this would be a user ID to identify the user connected.
		myClient = &myCustomClient{
			Client: ws.NewClient(h, conn, fn),
			id:     r.URL.Query().Get("id"),
		}

		return myClient
	}

	hub = ws.New(upgrader, clientFn)

	go func() {
		for {
			time.Sleep(time.Second * 4)

			hub.Broadcast([]byte("Server says Hi!"))
		}
	}()

	http.HandleFunc("/", root)
	http.HandleFunc("/ws", webs)

	http.ListenAndServe(":8080", nil)
}

func webs(w http.ResponseWriter, r *http.Request) {

	err := hub.Accept(w, r)
	if err != nil {
		log.Printf("accepting WebSocket connection '%s'\n", err)
	}
}

func root(w http.ResponseWriter, r *http.Request) {

	s := struct {
		Addr string
	}{
		Addr: "localhost:8080",
	}

	err := tpls.ExecuteTemplate(w, "root", s)
	if err != nil {
		log.Fatal(err)
	}
}

const (
	html = `<html>
	<head>
	</head>
	<body>
		<p>root</p>
		<div id="ct-main">
		</div>
		<script type="text/javascript">

			function getRandom(min, max) {
				return Math.floor(Math.random() * (max - min) + min);
			}

			var clientID = getRandom(1,1000),
				ct = document.getElementById('ct-main'),
			    conn = new WebSocket("ws://{{ .Addr }}/ws?id="+clientID);
			
			conn.onclose = function (evt) {

	            var item = document.createElement("div");
	            item.innerHTML = "<b>Connection closed.</b>";
	            ct.appendChild(item);
	        };

	        conn.onmessage = function (evt) {

	            var messages = evt.data.split('\n');

	            for (var i = 0; i < messages.length; i++) {
	                var item = document.createElement("div");
	                item.innerText = messages[i];
	                ct.appendChild(item);
	            }
	        };

			conn.onopen = function(evt){

				var i = 0;

				console.log("connection opened, client " + clientID);

				setInterval(function(){ i++; conn.send('Client ' + clientID + ' says hi '+i); }, 2000);
			};

		</script>
	</body>
</html>`
)
