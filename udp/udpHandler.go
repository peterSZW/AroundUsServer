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

	packetsQueue = goconcurrentqueue.NewFIFO()

	//Basic variables
	addresss := fmt.Sprintf("%s:%d", host, port)
	protocol := "udp"

	log.Println("Starting UDP listening", addresss)
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

	//一个 getIncomingUdp()，多个 handleIncomingUdpData()
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

		var dataPacket packet.ClientPacketRaw
		err = json.Unmarshal(dequeuedPacket.Data, &dataPacket)
		if err != nil {
			log.Println("Couldn't parse json player data! Skipping iteration!")
			continue
		}

		//解码出来数据

		err = handleUdpData(dequeuedPacket.Address, dataPacket, dequeuedPacket.Data)
		if err != nil {
			log.Println("Error while handling UDP packet: " + err.Error())
			continue
		}
	}
}

func handleUdpData(userAddress *net.UDPAddr, clientPacket packet.ClientPacketRaw, packetData []byte) error {
	//log.Println(clientPacket)

	log.Println("-<-", string(packetData))

	// log.Println("#1#", clientPacket)
	if clientPacket.Type == packet.DialAddr { // {"type":5, "id": 0}
		if user, ok := player.PlayerList[clientPacket.Uuid]; ok {
			user.UdpAddress = userAddress
		}
		return nil
	}

	// log.Println("#2#", clientPacket)
	// dataPacket, err := clientPacket.DataToBytes()
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }

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

		player1, err := UnmarshalUser([]byte(packetData))
		if err == nil {

			currUser := player1.InitializePlayer()

			{

				//defer deInitializePlayer(currUser)
				//TODO: defer notify all that player left
				//TODO: notify player about all players in lobby
				//TODO: notify all that player joined

				currUser.UdpAddress = userAddress

				player.PlayerListLock.Lock()

				player.PlayerList[currUser.Uuid] = currUser
				player.PlayerListLock.Unlock()

				for i, obj := range player.PlayerList {
					fmt.Println(i, "-", obj)
				}

			}
		} else {
			log.Println(err)
		}

	case packet.UserDisconnected: // example: {"type":12, "data":{"name":"bro", "color": 1}}

		player1, err := UnmarshalUser([]byte(packetData))
		if err == nil {
			log.Println("DeInitializePlayer", player1)
			player1.DeInitializePlayer()
		} else {
			log.Println(err)
		}

	case packet.UpdatePos: // {"type":6, "id": 0, "data":{"x":0, "y":2, "z":69}}
		var newPosition player.PlayerPosition
		err := json.Unmarshal(packetData, &newPosition)
		if err != nil {
			return fmt.Errorf("cant parse position player data")
		}
		playee := player.PlayerList[clientPacket.Uuid]
		if playee != nil {
			playee.PlayerPosition = newPosition
		}
		//player.PlayerList[clientPacket.Uuid].PlayerPosition = newPosition
	case packet.UpdateRotation: // {"type":7, "id": 0, "data":{"pitch":42, "yaw":11}}
		var newRotation player.PlayerRotation
		err := json.Unmarshal(packetData, &newRotation)
		if err != nil {
			return fmt.Errorf("cant parse rotation player data")
		}
		player.PlayerList[clientPacket.Uuid].Rotation = newRotation
	default:
		if user, ok := player.PlayerList[clientPacket.Uuid]; ok {
			tcp.SendErrorMsg(user.TcpConnection, "Invalid UDP packet type!")
		}

	}

	return nil
}

func updatePlayerPosition() {
	for {
		if len(player.PlayerList) > 1 {
			for _, user := range player.PlayerList {
				//TODO send name or id as well
				//BUG where only one recieves
				fmt.Println(user)
				BroadcastUDP(user, packet.PositionBroadcast, []string{user.Uuid})
			}
		}
		//log.Println("Loop position")
		time.Sleep(500 * time.Millisecond)
	}
}

// function wont send the message for players in the filter
func BroadcastUDP(data interface{}, packetType int8, userFilter []string) error {
	packetToSend := packet.StampPacket("", data, packetType)
	for _, user := range player.PlayerList {
		if !utils.IntInArray(user.Uuid, userFilter) && user.UdpAddress != nil {
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
