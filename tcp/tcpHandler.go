package tcp

import (
	"aroundUsServer/globals"
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"encoding/json"
	"net"
	"strconv"
	"strings"

	"github.com/inconshreveable/log15"
)

func ListenTCP(host string, port int) {
	tcpListener, err := net.Listen("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log15.Crit("Listen", "err", err)
	}
	log15.Error("Starting TCP listening")
	defer tcpListener.Close()

	for {
		tcpConnection, err := tcpListener.Accept()
		if err != nil {
			log15.Error("tcplisten", "err", err)
			continue
		}

		go handleTcpPlayer(tcpConnection)
	}
}

func handleTcpPlayer(conn net.Conn) {
	log15.Error("Accepted new connection.")
	defer conn.Close()
	defer log15.Error("Closed connection.")

	if !globals.IsInLobby {
		SendErrorMsg(conn, "Game has already started!")
		return
	}

	//var currUser *player.Player

	for {
		// Max packet is 1024 bytes long
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			SendErrorMsg(conn, "Error while reading the packet!\n"+err.Error())
			log15.Error(string(buf))
			return
		}
		rawStreamData := []byte(strings.TrimSpace(string(buf[:size])))

		log15.Error(string(rawStreamData))

		// Get the packet ID from the JSON
		var clientPacket packet.ClientPacket
		err = json.Unmarshal(rawStreamData, &clientPacket)
		if err != nil {
			log15.Error("Couldn't parse json player data! Skipping iteration! " + err.Error())
			continue
		}

		// packetData := []byte(clientPacket.Data)
		// packetData, ok :=  fmt.Sprint(data)
		// log15.Error(packetData)
		// if !ok {
		// 	log15.Error("Couldn't turn data to []byte! Skipping iteration! ")
		// 	continue
		// }

		jsonString, err := json.Marshal(clientPacket.Data)
		if err != nil {
			SendErrorMsg(conn, "Cant turn inteface to json!\n"+err.Error())
			continue
		}
		packetData := []byte(jsonString)

		// packetData, err := clientPacket.DataToBytes()
		// if err != nil {
		// 	log15.Error("Cant turn inteface to []byte!")
		// 	return
		// }
		// log15.Error(string(packetData))

		switch clientPacket.Type {
		case packet.NewUser: // example: {"type":1, "data":{"name":"bro", "color": 1}}
			_, err := initializePlayer([]byte(packetData), conn)
			if err != nil {
				SendErrorMsg(conn, "error while making a user: "+err.Error())
				return
			}

			//defer deInitializePlayer(currUser)

			//player.PlayerList[currUser.Uuid] = currUser

			// conenctedUsersJSON, err := json.Marshal(playerList) // Get all the players before adding the current user
			// if err != nil {
			// 	sendErrorMsg(conn, "Error while Marshaling the current connected users, disconnecting the user")
			// 	return
			// }
			// currUser.unduplicateUsername()

			// playerList[currUser.id] = currUser // Add the current user to the player map

			// defer currUser.removePlayer()

			// // Tell old users that a user connected
			// currUserJSON, err := json.Marshal(currUser) // Get all the players before adding the current user
			// if err != nil {
			// 	sendErrorMsg(conn, "Error while Marshaling the current user, other users dont know of your existance!")
			// }

			// currUserJSON, err = encapsulatePacketID(NewPlayerConnected, currUserJSON)
			// if err != nil {
			// 	log15.Error("Didn't encapsulate currUserJSON with ID")
			// }
			// sendEveryoneTcpData([]byte(currUserJSON), []string{currUser.Name})

			// // Tell the current user where to spawn at
			// ClientSpawnPositionJSON, err := json.Marshal(currUser.PlayerPosition) // Get all the players before adding the current user
			// if err != nil {
			// 	sendErrorMsg(conn, "Error while Marshaling the current user position")
			// }
			// ClientSpawnPositionJSON, err = encapsulatePacketID(ClientSpawnPosition, ClientSpawnPositionJSON)
			// if err != nil {
			// 	log15.Error("Didn't encapsulate currUserJSON with ID")
			// }
			// conn.Write([]byte(stampPacketLength(ClientSpawnPositionJSON)))

			// // Tell the current user who is already connected
			// conenctedUsersJSON, err = encapsulatePacketID(UsersInGame, conenctedUsersJSON)
			// if err != nil {
			// 	log15.Error("Didn't encapsulate currUserJSON with ID")
			// }
			// conn.Write([]byte(stampPacketLength(conenctedUsersJSON)))

			// // Tell the user if he is manager
			// conenctedUsersJSON, err = encapsulatePacketID(IsUserManager, []byte(strconv.FormatBool(currUser.isManager)))
			// if err != nil {
			// 	log15.Error("Didn't encapsulate currUserJSON with ID")
			// }
			// conn.Write([]byte(stampPacketLength(conenctedUsersJSON)))

			// // Tell the his ID
			// conenctedUsersJSON, err = encapsulatePacketID(UserId, []byte(strconv.FormatInt(int64(currUser.id), 10)))
			// if err != nil {
			// 	log15.Error("Didn't encapsulate currUserJSON with ID")
			// }
			// conn.Write([]byte(stampPacketLength(conenctedUsersJSON)))

		// case StartGame:
		// 	var rotation playerRotation
		// 	data, err := packet.dataToBytes()
		// 	if err != nil {
		// 		log15.Error("Cant turn inteface to []byte!")
		// 		return
		// 	}
		// 	err = json.Unmarshal(data, &rotation)
		// 	if err != nil {
		// 		log15.Error("Cant parse json init player data!")
		// 	}
		// 	playerList[currUser.id].Rotation = rotation.Rotation
		default:
			SendErrorMsg(conn, "Invalid packet type!")

		}

	}
}

func initializePlayer(data []byte, tcpConnection net.Conn) (*player.Player, error) {
	// player.PlayerListLock.Lock()
	// defer player.PlayerListLock.Unlock()

	var newPlayer *player.Player
	err := json.Unmarshal(data, &newPlayer)
	if err != nil {
		log15.Error("Cant parse json init player data!")
		return nil, err
	}

	newPlayer.TcpConnection = tcpConnection // Set the player TCP connection socket

	// check if the name is taken or invalid
	// we need to keep a counter so the name will be in the format `<name> <count>`
	var newNameCount int16

	oldName := newPlayer.Name
	// var nameOk bool
	// for !nameOk {
	// 	nameOk = true
	// 	for _, player := range player.PlayerList {
	// 		if player.Name == newPlayer.Name {
	// 			newNameCount++
	// 			nameOk = false
	// 			newPlayer.Name = fmt.Sprintf("%s %d", oldName, newNameCount)
	// 			break
	// 		}
	// 	}
	// }

	if newNameCount == 0 {
		newPlayer.Name = oldName
	}

	// check if the color is taken or invalid, if it is assign next not taken color
	if int16(0) > newPlayer.Color || int16(len(player.Colors)) <= newPlayer.Color || player.Colors[newPlayer.Color] {
		for index, color := range player.Colors {
			if !color {
				newPlayer.Color = int16(index)
				break
			}
		}
	}

	player.Colors[newPlayer.Color] = true // set player color as taken

	// check if he is the first one in the lobby, if true set the player to be the game manager
	// if len(player.PlayerList) == 0 {
	// 	newPlayer.IsManager = true
	// }

	// set player ID and increase to next one, theoretically this can roll back at 2^31-1
	newPlayer.Uuid = strconv.Itoa(player.CurrId)
	player.CurrId++

	// set player spawn position
	newPlayer.PlayerPosition = player.SpawnPositionsStack[len(player.SpawnPositionsStack)-1]  // peek at the last element
	player.SpawnPositionsStack = player.SpawnPositionsStack[:len(player.SpawnPositionsStack)] // pop

	log15.Error("New player got generated:")
	newPlayer.PrintUser()

	return newPlayer, nil
}

// func deInitializePlayer(playerToDelete *player.Player) error {
// 	player.PlayerListLock.Lock()
// 	defer player.PlayerListLock.Unlock()

// 	delete(player.PlayerList, playerToDelete.Uuid)

// 	// give another player the manager
// 	if playerToDelete.IsManager {
// 		for _, nextPlayer := range player.PlayerList {
// 			nextPlayer.IsManager = true
// 			break
// 		}
// 	}

// 	// free the color
// 	player.Colors[playerToDelete.Color] = false

// 	playerToDelete = nil

// 	return nil
// }

func SendErrorMsg(conn net.Conn, msg string) error {
	log15.Error(msg)
	errorPacket := packet.StampPacket("", []byte(msg), packet.Error)
	_, err := errorPacket.SendTcpStream(conn)
	return err
}

// function wont send the message for players in the filter
func BroadcastTCP(data []byte, packetType int16, userFilter []string) error {
	// for _, user := range player.PlayerList {
	// 	if !utils.IntInArray(user.Uuid, userFilter) {
	// 		log15.Error("Sending data to everyone(Filtered) " + string(data))
	// 		packetToSend := packet.StampPacket("", data, packetType)
	// 		_, err := packetToSend.SendTcpStream(user.TcpConnection)
	// 		if err != nil {
	// 			log15.Error("SendTcpStream", "err", err)
	// 		}
	// 	}
	// }
	return nil
}
