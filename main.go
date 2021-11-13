package main

import (
	"aroundUsServer/cli"
	"aroundUsServer/player"
	"aroundUsServer/udp"
	"flag"
	"io/ioutil"

	//"log"
	"strings"

	"github.com/inconshreveable/log15"
)

/*
** DISCLAIMER! **
This server is not designed to check the users inputs!
This server is ~quick~ and dirty to be able to play with friends a game and most calculations
get calculated in the client so the server is highly trusting the clients.
Its not designed to be released to the wild and shouldn't be trusted with random users.
If you use this server in the wild, cheating & crashing will be SO ez.
The unity game client I built wont be released as I respect the developers of "Among Us".
*/

var host *string
var port *int
var userList map[string]string

func main() {
	// init variables

	// get program flags
	host = flag.String("ip", "0.0.0.0", "Server listen IP")
	port = flag.Int("port", 7403, "Server listen port")
	var isclient = flag.Bool("client", false, "client")
	flag.Parse()

	data, err := ioutil.ReadFile("user.txt")
	if err != nil {
		log15.Error("ReadFile", err)
	}
	userlist := string(data)
	strings.Split(userlist, "\n")

	if *isclient {
		log15.Debug("Starting client")

		client()

	} else {

		initSpawnPosition()

		go start_websocket_server()

		// start listening
		//go tcp.ListenTCP(*host, *port)
		go udp.ListenUDP(*host, *port)

		// block main thread with the console
		cli.ServerConsoleCLI()

	}
}

// func (p player.Player) isInFilter(filter []string) bool {
// 	for _, name := range filter {
// 		if name == p.Name {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (p player.Player) unduplicateUsername() {
// 	var nextNumber int16
// 	wasDuped := true
// 	criticalUseLock.Lock()
// 	for wasDuped {
// 		wasDuped = false
// 		for _, player := range playerList {
// 			if player.Name == p.Name {
// 				nextNumber++
// 				wasDuped = true
// 				break
// 			}
// 		}
// 	}
// 	criticalUseLock.Unlock()
// 	if nextNumber != 0 {
// 		p.Name = p.Name + strconv.Itoa(int(nextNumber))
// 	}
// }

// func (p player) removePlayer() {
// 	criticalUseLock.Lock()
// 	delete(playerList, p.id)
// 	criticalUseLock.Unlock()

// 	currUserJSON, err := json.Marshal(p) // Get all the players before adding the current user
// 	if err != nil {
// 		sendErrorMsg(p.tcpConnection, "Error while Marshaling the user for remove, brotha tell ofido!")
// 		return
// 	}

// 	currUserJSON, err = encapsulatePacketID(UserDisconnected, currUserJSON)
// 	if err != nil {
// 		log15.Error("Didn't encapsulate currUserJSON with ID")
// 		return
// 	}
// 	sendEveryoneTcpData([]byte(currUserJSON), []string{p.Name})
// }

func initSpawnPosition() {
	for i := 5; i <= 0; i++ {
		player.SpawnPositionsStack = append(player.SpawnPositionsStack, player.PlayerPosition{X: -4, Y: 1.75, Z: float32(14 - i)})
		player.SpawnPositionsStack = append(player.SpawnPositionsStack, player.PlayerPosition{X: -6, Y: 1.75, Z: float32(14 - i)})
	}
}

// func encapsulatePacketID(ID int, JSON []byte) ([]byte, error) {
// 	errorJSON, err := json.Marshal(packetType{ID, JSON})
// 	return errorJSON, err
// }

// func stampPacketLength(data []byte) []byte {
// 	packet := make([]byte, 0, 4+len(data))
// 	packet = append(packet, []byte(fmt.Sprintf("%04d", len(data)))...)
// 	packet = append(packet, data...)
// 	return packet
// }
