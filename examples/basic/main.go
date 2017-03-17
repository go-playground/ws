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

func main() {

	var err error

	tpls, err = template.New("root").Parse(html)
	if err != nil {
		log.Fatal(err)
	}

	hub = ws.New(upgrader, nil)

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
