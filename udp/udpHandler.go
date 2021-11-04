package udp

import (
	"aroundUsServer/globals"
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"aroundUsServer/tcp"
	"aroundUsServer/utils"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
)

var packetsQueue *goconcurrentqueue.FIFO
var udpConnection *net.UDPConn

type udpPacket struct {
	Address *net.UDPAddr
	Data    []byte
}

func ListenUDP(host string, port int) {
	log.Println("Starting UDP listening")

	packetsQueue = goconcurrentqueue.NewFIFO()

	//Basic variables
	addresss := fmt.Sprintf("%s:%d", host, port)
	protocol := "udp"

	//Build the address
	udpAddr, err := net.ResolveUDPAddr(protocol, addresss)
	if err != nil {
		log.Println("Wrong Address")
		return
	}

	//Create the connection
	udpConnection, err = net.ListenUDP(protocol, udpAddr)
	if err != nil {
		log.Println(err)
	}

	// create queue readers
	for i := 0; i < globals.QueueReaders; i++ {
		go handleIncomingUdpData()
	}

	// reate position updater
	go updatePlayerPosition()

	//Keep calling this function
	for {
		quit := make(chan struct{})
		for i := 0; i < 1; i++ {
			go getIncomingUdp(quit)
		}
		<-quit // hang until an error
	}
}

func getIncomingUdp(quit chan struct{}) {
	err := error(nil)

	for err == nil {
		buffer := make([]byte, 1024)

		size, addr, err := udpConnection.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Cant read packet!", err)
			continue
		}
		data := buffer[:size]

		packetsQueue.Enqueue(udpPacket{Address: addr, Data: data})
	}

	log.Println("Listener failed - restarting!", err)
	quit <- struct{}{}
}

func handleIncomingUdpData() {
	for {
		dequeuedRawPacket, err := packetsQueue.DequeueOrWaitForNextElement()
		if err != nil {
			log.Println("Couldn't dequeue!")
			continue
		}

		dequeuedPacket, ok := dequeuedRawPacket.(udpPacket)
		if !ok {
			log.Println("Couldn't turn udp data to udpPacket!")
			continue
		}

		var dataPacket packet.ClientPacket
		err = json.Unmarshal(dequeuedPacket.Data, &dataPacket)
		if err != nil {
			log.Println("Couldn't parse json player data! Skipping iteration!")
			continue
		}

		err = handleUdpData(dequeuedPacket.Address, dataPacket, dequeuedPacket.Data)
		if err != nil {
			log.Println("Error while handling UDP packet: " + err.Error())
			continue
		}
	}
}

func handleUdpData(userAddress *net.UDPAddr, clientPacket packet.ClientPacket, packetData []byte) error {
	//log.Println(clientPacket)

	log.Println(string(packetData))

	if clientPacket.Type == packet.DialAddr { // {"type":5, "id": 0}
		if user, ok := globals.PlayerList[clientPacket.PlayerID]; ok {
			user.UdpAddress = userAddress
		}
		return nil
	}

	dataPacket, err := clientPacket.DataToBytes()
	if err != nil {
		return err
	}

	switch clientPacket.Type {

	/*
		InitUser       = iota + 1 // TCP
		KilledPlayer              // TCP
		GameInit                  // TCP
		StartGame                 // TCP
		DialAddr                  // UDP
		UpdatePos                 // UDP
		UpdateRotation            // UDP
	*/

	case packet.InitUser: // example: {"type":1, "data":{"name":"bro", "color": 1}}

		player, err := UnmarshalUser([]byte(packetData))
		if err == nil {

			currUser := initializePlayer(player)

			{

				//defer deInitializePlayer(currUser)
				//TODO: defer notify all that player left
				//TODO: notify player about all players in lobby
				//TODO: notify all that player joined

				currUser.UdpAddress = userAddress

				globals.PlayerListLock.Lock()

				globals.PlayerList[currUser.Id] = currUser
				globals.PlayerListLock.Unlock()

				for i, obj := range globals.PlayerList {
					fmt.Println(i, "-", obj)
				}

			}
		} else {
			log.Println(err)
		}

	case packet.UpdatePos: // {"type":6, "id": 0, "data":{"x":0, "y":2, "z":69}}
		var newPosition player.PlayerPosition
		err := json.Unmarshal([]byte(dataPacket), &newPosition)
		if err != nil {
			return fmt.Errorf("cant parse position player data")
		}
		globals.PlayerList[clientPacket.PlayerID].PlayerPosition = newPosition
	case packet.UpdateRotation: // {"type":7, "id": 0, "data":{"pitch":42, "yaw":11}}
		var newRotation player.PlayerRotation
		_ = json.Unmarshal([]byte(dataPacket), &newRotation)
		if err != nil {
			return fmt.Errorf("cant parse rotation player data")
		}
		globals.PlayerList[clientPacket.PlayerID].Rotation = newRotation
	default:
		if user, ok := globals.PlayerList[clientPacket.PlayerID]; ok {
			tcp.SendErrorMsg(user.TcpConnection, "Invalid UDP packet type!")
		}

	}

	return nil
}

func updatePlayerPosition() {
	for {
		if len(globals.PlayerList) > 1 {
			for _, user := range globals.PlayerList {
				//TODO send name or id as well
				//BUG where only one recieves
				fmt.Println(user)
				BroadcastUDP(user, packet.PositionBroadcast, []int{user.Id})
			}
		}
		//log.Println("Loop position")
		time.Sleep(500 * time.Millisecond)
	}
}

// function wont send the message for players in the filter
func BroadcastUDP(data interface{}, packetType int8, userFilter []int) error {
	packetToSend := packet.StampPacket(data, packetType)
	for _, user := range globals.PlayerList {
		if !utils.IntInArray(user.Id, userFilter) && user.UdpAddress != nil {
			_, err := packetToSend.SendUdpStream(udpConnection, user.UdpAddress)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}

func UnmarshalUser(data []byte) (*player.Player, error) {
	type Tdata struct {
		Type int8           `json:"type"`
		Seq  int64          `json:"seq"`
		Data *player.Player `json:"data"`
	}
	var dataobj Tdata

	err := json.Unmarshal(data, &dataobj)
	if err != nil {
		log.Println("Cant parse json init player data!")
		return nil, err
	}
	return dataobj.Data, nil

}

func initializePlayer(newPlayer *player.Player) *player.Player {

	log.Println("===========", newPlayer)

	//newPlayer.TcpConnection = tcpConnection // Set the player TCP connection socket

	// check if the name is taken or invalid
	// we need to keep a counter so the name will be in the format `<name> <count>`

	var newNameCount int8
	var nameOk bool
	oldName := newPlayer.Name

	for !nameOk {
		nameOk = true
		for _, player := range globals.PlayerList {
			if player.Name == newPlayer.Name {
				newNameCount++
				nameOk = false
				newPlayer.Name = fmt.Sprintf("%s %d", oldName, newNameCount)
				break
			}
		}
	}

	if newNameCount == 0 {
		newPlayer.Name = oldName
	}

	// check if the color is taken or invalid, if it is assign next not taken color
	if int8(0) > newPlayer.Color || int8(len(globals.Colors)) <= newPlayer.Color || globals.Colors[newPlayer.Color] {
		for index, color := range globals.Colors {
			if !color {
				newPlayer.Color = int8(index)
				break
			}
		}
	}

	globals.Colors[newPlayer.Color] = true // set player color as taken

	// check if he is the first one in the lobby, if true set the player to be the game manager
	if len(globals.PlayerList) == 0 {
		newPlayer.IsManager = true
	}

	// set player ID and increase to next one, theoretically this can roll back at 2^31-1
	newPlayer.Id = globals.CurrId
	globals.CurrId++

	// set player spawn position
	newPlayer.PlayerPosition = globals.SpawnPositionsStack[len(globals.SpawnPositionsStack)-1]   // peek at the last element
	globals.SpawnPositionsStack = globals.SpawnPositionsStack[:len(globals.SpawnPositionsStack)] // pop

	log.Println("New player got generated:")
	utils.PrintUser(newPlayer)

	return newPlayer
}

func deInitializePlayer(playerToDelete *player.Player) error {
	globals.PlayerListLock.Lock()
	defer globals.PlayerListLock.Unlock()

	delete(globals.PlayerList, playerToDelete.Id)

	// give another player the manager
	if playerToDelete.IsManager {
		for _, nextPlayer := range globals.PlayerList {
			nextPlayer.IsManager = true
			break
		}
	}

	// free the color
	globals.Colors[playerToDelete.Color] = false

	playerToDelete = nil

	return nil
}
