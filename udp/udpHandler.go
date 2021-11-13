package udp

import (
	"aroundUsServer/globals"
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"aroundUsServer/tcp"
	"aroundUsServer/utils"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
	"github.com/xiaomi-tc/log15"
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

	log15.Error("Starting UDP listening", addresss)
	//Build the address
	udpAddr, err := net.ResolveUDPAddr(protocol, addresss)
	if err != nil {
		log15.Error("Wrong Address")
		return
	}

	//Create the connection
	udpConnection, err = net.ListenUDP(protocol, udpAddr)
	if err != nil {
		log15.Error("ListenUDP", err)
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
			log15.Error("Cant read packet!", err)
			continue
		}
		data := buffer[:size]

		packetsQueue.Enqueue(udpPacket{Address: addr, Data: data})
	}

	log15.Error("Listener failed - restarting!", err)
	quit <- struct{}{}
}

func handleIncomingUdpData() {
	for {
		dequeuedRawPacket, err := packetsQueue.DequeueOrWaitForNextElement()
		if err != nil {
			log15.Error("Couldn't dequeue!")
			continue
		}

		dequeuedPacket, ok := dequeuedRawPacket.(udpPacket)
		if !ok {
			log15.Error("Couldn't turn udp data to udpPacket!")
			continue
		}

		var dataPacket packet.ClientPacketRaw
		err = json.Unmarshal(dequeuedPacket.Data, &dataPacket)
		if err != nil {
			log15.Error("Couldn't parse json player data! Skipping iteration!")
			continue
		}

		//解码出来数据

		err = handleUdpData(dequeuedPacket.Address, dataPacket, dequeuedPacket.Data)
		if err != nil {
			log15.Error("Error while handling UDP packet: " + err.Error())
			continue
		}
	}
}

func handleUdpData(userAddress *net.UDPAddr, clientPacket packet.ClientPacketRaw, packetData []byte) error {
	//log15.Error(clientPacket)

	log15.Error("-<-", string(packetData))

	// log15.Error("#1#", clientPacket)
	if clientPacket.Type == packet.DialAddr { // {"type":5, "id": 0}
		if user, ok := player.PlayerList[clientPacket.Uuid]; ok {
			user.UdpAddress = userAddress
		}
		return nil
	}

	// log15.Error("#2#", clientPacket)
	// dataPacket, err := clientPacket.DataToBytes()
	// if err != nil {
	// 	log15.Error(err)
	// 	return err
	// }

	switch clientPacket.Type {

	/*
		InitUser            = iota + 1 // TCP// Client -> Server packets
		KilledPlayer                   // TCP
		GameInit                       // TCP
		StartGame                      // TCP
		DialAddr                       // UDP
		UpdatePos                      // UDP
		UpdateRotation                 // UDP
		HeartBeat                      // UDP
		UsersInGame                    // TCP// Server -> Client packets
		IsUserManager                  // TCP
		NewPlayerConnected             // TCP
		ClientSpawnPosition            // TCP
		UserDisconnected               // TCP
		GameOver                       // TCP
		PlayerDied                     // TCP
		UserId                         // TCP
		Error                          // TCP
		PositionBroadcast              // UDP
	*/

	case packet.NewUser: // example: {"type":1, "data":{"name":"bro", "color": 1}}

		var dataobj packet.TNewUserReq
		err := json.Unmarshal(packetData, &dataobj)

		if err != nil {
			log15.Error("Cant parse json init player data!")
		} else {
			dataobj.Data.Uuid = dataobj.Uuid
			player1 := dataobj.Data

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
					fmt.Println("(", i, ")", obj)
				}

			}
		}

	case packet.Disconnect: // example: {"type":12, "data":{"name":"bro", "color": 1}}

		var dataobj packet.TDisconnectReq
		err := json.Unmarshal(packetData, &dataobj)

		if err == nil {
			log15.Debug("DeInitializePlayer", dataobj)
			//player1.DeInitializePlayer()
		} else {
			log15.Error("Unmarshal", err)
		}

	case packet.UpdatePos: // {"type":6, "id": 0, "data":{"x":0, "y":2, "z":69}}
		var dataobj packet.TUpdatePosReq
		err := json.Unmarshal(packetData, &dataobj)
		if err != nil {
			return fmt.Errorf("cant parse position player data")
		}
		playee := player.PlayerList[clientPacket.Uuid]
		if playee != nil {
			playee.PlayerPosition = dataobj.PP
		}
		//player.PlayerList[clientPacket.Uuid].PlayerPosition = newPosition
	case packet.UpdateRotation: // {"type":7, "id": 0, "data":{"pitch":42, "yaw":11}}
		var dataobj packet.TUpdateRotationReq
		err := json.Unmarshal(packetData, &dataobj)
		if err != nil {
			return fmt.Errorf("cant parse rotation player data")
		}
		player.PlayerList[clientPacket.Uuid].Rotation = dataobj.PP
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
		//log15.Error("Loop position")
		time.Sleep(500 * time.Millisecond)
	}
}

// function wont send the message for players in the filter
func BroadcastUDP(data interface{}, packetType int16, userFilter []string) error {
	packetToSend := packet.StampPacket("", data, packetType)
	for _, user := range player.PlayerList {
		if !utils.IntInArray(user.Uuid, userFilter) && user.UdpAddress != nil {
			_, err := packetToSend.SendUdpStream(udpConnection, user.UdpAddress)
			if err != nil {
				log15.Error("SendUdpStream", err)
			}
		}
	}
	return nil
}

func UnmarshalUser(data []byte) (*player.Player, error) {
	type Tdata struct {
		Type int16          `json:"type"`
		Seq  int64          `json:"seq"`
		Data *player.Player `json:"data"`
	}
	var dataobj Tdata

	err := json.Unmarshal(data, &dataobj)
	if err != nil {
		log15.Error("Cant parse json init player data!")
		return nil, err
	}
	return dataobj.Data, nil

}
