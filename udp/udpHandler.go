package udp

import (
	"aroundUsServer/globals"
	"aroundUsServer/packet"
	"aroundUsServer/player"
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
	go RemoveUnActivePlayer(time.Second)

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

		handleUdpData(dequeuedPacket.Address, dataPacket, dequeuedPacket.Data)

	}
}

func handleUdpData(userAddress *net.UDPAddr, clientPacket packet.TBaseReqPacket, packetData []byte) {

	log15.Debug("-<-", "packet", string(packetData))

	if clientPacket.Type == packet.DialAddr { // {"type":5, "id": 0}
		aplayer, ok := player.PlayerMap.Load(clientPacket.Uuid)
		if ok {
			log15.Debug("SET UdpAddress", "Dial", userAddress)
			aplayer.(*player.Player).UdpAddress = userAddress
			aplayer.(*player.Player).LastUpdate = time.Now()

		} else {
			log15.Debug("Not Found", "uuid", clientPacket.Uuid)

		}
		return

	}

	switch clientPacket.Type {

	//===================================================================
	//===================================================================
	//===================================================================
	//===================================================================

	case packet.UpdatePos: // {"type":6, "id": 0, "data":{"x":0, "y":2, "z":69}}

		var dataobj packet.TUpdatePosReq
		err := json.Unmarshal(packetData, &dataobj)
		if err != nil {
			log15.Debug("TUpdatePosReq Unmarshal", "err", err)
			return
			//return fmt.Errorf("cant parse position player data")
		}

		aplayer, ok := player.PlayerMap.Load(dataobj.Uuid)
		if ok {
			log15.Debug("SET PlayerPosition", "pos", dataobj.Data)
			aplayer.(*player.Player).LastUpdate = time.Now()
			aplayer.(*player.Player).PlayerPosition = dataobj.Data
			updateOnePlayerPositionNow(aplayer.(*player.Player))
		} else {
			log15.Warn("Not Found", "uuid", dataobj.Uuid)

		}
		//===================================================================
		//===================================================================
		//===================================================================
		//===================================================================

	case packet.HeartBeat: // {"type":6, "id": 0, "data":{"x":0, "y":2, "z":69}}

		var dataobj packet.THeartBeatReq
		err := json.Unmarshal(packetData, &dataobj)
		if err != nil {
			log15.Error("THeartBeatReq Unmarshal", "err", err)
			return
		}
		aplayer, ok := player.PlayerMap.Load(dataobj.Uuid)
		if ok {

			aplayer.(*player.Player).LastUpdate = time.Now()

		} else {
			log15.Warn("Not Found", "uuid", dataobj.Uuid)

		}

		//===================================================================
		//===================================================================
		//===================================================================
		//===================================================================

	case packet.UpdateRotation: // {"type":7, "id": 0, "data":{"pitch":42, "yaw":11}}
		var dataobj packet.TUpdateRotationReq
		err := json.Unmarshal(packetData, &dataobj)
		if err != nil {
			log15.Error("TUpdateRotationReq Unmarshal", "err", err)
			return

		}

		aplayer, ok := player.PlayerMap.Load(dataobj.Uuid)
		if ok {
			log15.Debug("SET Rotation", "Rotation", dataobj.Data)
			aplayer.(*player.Player).LastUpdate = time.Now()
			aplayer.(*player.Player).Rotation = dataobj.Data
			updateOnePlayerPositionNow(aplayer.(*player.Player))
		} else {
			log15.Warn("Not Found", "uuid", dataobj.Uuid)

		}
		//===================================================================
		//===================================================================
		//===================================================================
		//===================================================================

	default:
		log15.Warn("Unknow Type", "clientPacket", clientPacket)

	}

}

func updateOnePlayerPositionNow(user *player.Player) {

	BroadcastUDP(user, packet.PositionBroadcast, []string{user.Uuid})

}

func RemoveUnActivePlayer(sleepTime time.Duration) {

	for {
		player.PlayerMap.Range(func(k, v interface{}) bool {
			user := v.(*player.Player)
			if user.LastUpdate.Add(5 * time.Minute).Before(time.Now()) {
				player.PlayerMap.Delete(user.Uuid)
			}

			return true
		})
		time.Sleep(sleepTime)
	}

}

func updatePlayerPositionNow() {

	player.PlayerMap.Range(func(k, v interface{}) bool {
		user := v.(*player.Player)
		BroadcastUDP(user, packet.PositionBroadcast, []string{user.Uuid})

		return true
	})

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
