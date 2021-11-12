package main

import (
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

		jsonStu := HandleMessage2(message)

		// var msg MessageObj

		// err = json.Unmarshal(message, &msg)
		// if err != nil {
		// 	msg.Code = 1
		// 	msg.Data = string(message)

		// }
		// jsonStu, _ := json.Marshal(msg)

		err = c.WriteMessage(mt, []byte(jsonStu))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func HandleMessage2(packetData []byte) string {
	var dataobj packet.TBaseReqPacket
	err := json.Unmarshal(packetData, &dataobj)
	rsp := packet.TBaseRspPacket{}
	if err != nil {

		rsp.Code = 500
		rsp.Msg = "Couldn't parse json data."
		rsp.MsgEx = err.Error()
		msg, _ := json.Marshal(rsp)
		return string(msg)

	}
	switch dataobj.Type {
	case packet.NewUser:
		var dataobj packet.TNewUserReq
		err := json.Unmarshal(packetData, &dataobj)

		if err != nil {
			log.Println("Cant parse json init player data!")
		} else {
			dataobj.Data.Uuid = dataobj.Uuid
			player1 := dataobj.Data

			currUser := player1.InitializePlayer()

			{

				player.PlayerListLock.Lock()
				player.PlayerList[currUser.Uuid] = currUser
				player.PlayerListLock.Unlock()

				for i, obj := range player.PlayerList {
					fmt.Println("(", i, ")", obj)
				}

			}
		}
	default:
		rsp.Code = 500
		rsp.Msg = "Unknow Type. " + strconv.Itoa(int(dataobj.Type))
		//rsp.MsgEx = err.Error()

	}

	msg, _ := json.Marshal(rsp)
	return string(msg)

}
func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

/*
curl -i -H "Content-Type: application/json" -X POST -d '{"user_id": "123", "coin":100, "success":1, "msg":"OK!" }'  http://127.0.0.1:7403/api
*/

func api(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	body_str := string(body)
	fmt.Println(body_str)
	result := HandleMessage2(body)
	w.Header().Set("content-type", "text/json")
	fmt.Fprint(w, string(result))

}

func start_websocket_server() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	http.HandleFunc("/api", api)
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