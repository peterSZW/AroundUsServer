package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "0.0.0.0:7403", "http service address")

var upgrader = websocket.Upgrader{} // use default options

type MessageObj struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg,omitempty"`
	MsgEx string      `json:"msgex,omitempty"`
	Data  interface{} `json:"data"`
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	go func() {
		for {
			msg := MessageObj{Code: 0, Data: time.Now()}
			msgbyte, _ := json.Marshal(msg)
			err = c.WriteMessage(1, msgbyte)
			if err != nil {
				log.Println("timmer write err:", err)
				break
			}
			time.Sleep(time.Second)
		}

	}()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		if mt != websocket.TextMessage {
			continue
		}
		log.Printf("recv: %s,%d", message, mt)

		var msg MessageObj

		err = json.Unmarshal(message, &msg)
		if err != nil {
			msg.Code = 1
			msg.Data = string(message)

		}
		jsonStu, _ := json.Marshal(msg)

		err = c.WriteMessage(mt, jsonStu)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func start_websocket_server() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Printf("Starting WSK listening %s:%d", *host, *port)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Println(err)
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
	var line_cnt=0;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
		line_cnt++;
		if (line_cnt>20) {
			line_cnt=0;
			output.innerHTML="";
		}
        output.appendChild(d);
		
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
		var packJson = {"code":0, "data":input.value};

		var jsonstr =JSON.stringify(packJson );//input.value

        print("SEND: " + jsonstr);

        ws.send(jsonstr);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
