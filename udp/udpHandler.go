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
	"github.com/inconshreveable/log15"
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

	log15.Debug("Starting UDP listening", "addr", addresss)
	//Build the address
	udpAddr, err := net.ResolveUDPAddr(protocol, addresss)
	if err != nil {
		log15.Error("Wrong Address")
		return
	}

	//Create the connection
	udpConnection, err = net.ListenUDP(protocol, udpAddr)
	if err != nil {
		log15.Error("ListenUDP", "err", err)
	}

	// create queue readers
	for i := 0; i < globals.QueueReaders; i++ {
		go handleIncomingUdpData()
	}

	// reate position updater
	go updatePlayerPositionEveryHalfSec()

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
			log15.Error("Cant read packet!", "err", err)
			continue
		}
		data := buffer[:size]

		packetsQueue.Enqueue(udpPacket{Address: addr, Data: data})
	}

	log15.Error("Listener failed - restarting!", "err", err)
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

		var dataPacket packet.TBaseReqPacket
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

func handleUdpData(userAddress *net.UDPAddr, clientPacket packet.TBaseReqPacket, packetData []byte) error {
	//log15.Error(clientPacket)

	log15.Debug("-<-", "packet", string(packetData))

	// log15.Error("#1#", clientPacket)
	if clientPacket.Type == packet.DialAddr { // {"type":5, "id": 0}
		aplayer, ok := player.PlayerMap.Load(clientPacket.Uuid)
		if ok {
			log15.Debug("SET UdpAddress", "Dial", userAddress)
			aplayer.(*player.Player).UdpAddress = userAddress
		} else {
			log15.Debug("Not Found", "uuid", clientPacket.Uuid)

		}

		// if user, ok := player.PlayerList[clientPacket.Uuid]; ok {
		// }
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

	case packet.UpdatePos: // {"type":6, "id": 0, "data":{"x":0, "y":2, "z":69}}

		var dataobj packet.TUpdatePosReq
		err := json.Unmarshal(packetData, &dataobj)
		if err != nil {
			return fmt.Errorf("cant parse position player data")
		}

		aplayer, ok := player.PlayerMap.Load(dataobj.Uuid)
		if ok {
			log15.Debug("SET PlayerPosition", "pos", dataobj.Data)
			aplayer.(*player.Player).PlayerPosition = dataobj.Data
			updateOnePlayerPositionNow(aplayer.(*player.Player))
		} else {
			log15.Debug("Not Found", "uuid", dataobj.Uuid)

		}
	case packet.UpdateRotation: // {"type":7, "id": 0, "data":{"pitch":42, "yaw":11}}
		var dataobj packet.TUpdateRotationReq
		err := json.Unmarshal(packetData, &dataobj)
		if err != nil {
			return fmt.Errorf("cant parse rotation player data")
		}

		aplayer, ok := player.PlayerMap.Load(dataobj.Uuid)
		if ok {
			log15.Debug("SET Rotation", "Rotation", dataobj.Data)
			aplayer.(*player.Player).Rotation = dataobj.Data
			updateOnePlayerPositionNow(aplayer.(*player.Player))
		} else {
			log15.Debug("Not Found", "uuid", dataobj.Uuid)

		}
	default:
		if user, ok := player.PlayerList[clientPacket.Uuid]; ok {
			tcp.SendErrorMsg(user.TcpConnection, "Invalid UDP packet type!")
		}

	}

	return nil
}

func updateOnePlayerPositionNow(user *player.Player) {

	BroadcastUDP(user, packet.PositionBroadcast, []string{user.Uuid})

	// if len(player.PlayerList) > 1 {
	// 	for _, user := range player.PlayerList {

	// 		BroadcastUDP(user, packet.PositionBroadcast, []string{user.Uuid})
	// 	}
	// }

}

func updatePlayerPositionNow() {

	player.PlayerMap.Range(func(k, v interface{}) bool {
		user := v.(*player.Player)
		BroadcastUDP(user, packet.PositionBroadcast, []string{user.Uuid})

		return true
	})

	// if len(player.PlayerList) > 1 {
	// 	for _, user := range player.PlayerList {

	// 		BroadcastUDP(user, packet.PositionBroadcast, []string{user.Uuid})
	// 	}
	// }

}

func updatePlayerPositionEveryHalfSec() {
	for {
		updatePlayerPositionNow()

		time.Sleep(5000 * time.Millisecond)
	}
}

// function wont send the message for players in the filter

func BroadcastUDP(data interface{}, packetType int16, userFilter []string) error {

	packetToSend := packet.StampPacket("", data, packetType)
	player.PlayerMap.Range(func(k, v interface{}) bool {
		user := v.(*player.Player)
		if !utils.IntInArray(user.Uuid, userFilter) && user.UdpAddress != nil {
			_, err := packetToSend.SendUdpStream(udpConnection, user.UdpAddress)
			if err != nil {
				log15.Error("SendUdpStream", "err", err)
			}
		}

		return true
	})

	return nil
}

func BroadcastUDP_Old(data interface{}, packetType int16, userFilter []string) error {
	packetToSend := packet.StampPacket("", data, packetType)
	for _, user := range player.PlayerList {
		if !utils.IntInArray(user.Uuid, userFilter) && user.UdpAddress != nil {
			_, err := packetToSend.SendUdpStream(udpConnection, user.UdpAddress)
			if err != nil {
				log15.Error("SendUdpStream", "err", err)
			}
		}
	}
	return nil
}

// func UnmarshalUser(data []byte) (*player.Player, error) {
// 	type Tdata struct {
// 		Type int16          `json:"type"`
// 		Seq  int64          `json:"seq"`
// 		Data *player.Player `json:"data"`
// 	}
// 	var dataobj Tdata

// 	err := json.Unmarshal(data, &dataobj)
// 	if err != nil {
// 		log15.Error("Cant parse json init player data!")
// 		return nil, err
// 	}
// 	return dataobj.Data, nil

// }
