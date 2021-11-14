package main

import (
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/imroc/req"
	"github.com/inconshreveable/log15"
)

func getIncomingClientUdp(udpConnection *net.UDPConn) {
	err := error(nil)
	fmt.Println("Client listen....")

	for err == nil {
		buffer := make([]byte, 1024)

		size, addr, err := udpConnection.ReadFromUDP(buffer)
		if err != nil {
			log15.Error("Cant read packet!", "err", err)
			continue
		}
		log15.Debug("ReadFromUDP", "addr", addr)
		data := buffer[:size]

		var dataPacket packet.ClientPacket
		err = json.Unmarshal(data, &dataPacket)
		if err != nil {
			log15.Error("Couldn't parse json player data! Skipping iteration!")
			continue
		} else {
			fmt.Println(dataPacket)
		}

	}

}

var user player.Player

func client() {

	udpAddr, _ := net.ResolveUDPAddr("udp4", *host+":"+strconv.Itoa(*port))
	// user.UdpAddress = &net.UDPAddr{
	// 	IP:   net.IPv4(127, 0, 0, 1),
	// 	Port: 7403,
	// }
	user.UdpAddress = udpAddr

	for {

		udpConnection, err := net.DialUDP("udp", nil, user.UdpAddress)

		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Duration(time.Second))

			continue
		}
		go getIncomingClientUdp(udpConnection)
		ClientConsoleCLI(udpConnection)
	}

}

var G_PlayerUuid1 string
var G_PlayerUuid2 string
var G_PlayerUuid3 string

func init() {
	G_PlayerUuid1 = NewUUID()
	G_PlayerUuid2 = NewUUID()
	G_PlayerUuid3 = NewUUID()

}

func ClientConsoleCLI(udpConnection *net.UDPConn) {

	for {
		var command, parameter string
		fmt.Scanln(&command, &parameter)

		switch command {

		case "n1":
			reqData := packet.TNewUserReq{Phone: "12"}
			reqData.Uuid = G_PlayerUuid1
			reqData.Type = packet.NewUser
			reqData.Data = &player.Player{Uuid: reqData.Uuid}

			data, _ := req.Post("http://127.0.0.1:7403/api", req.BodyJSON(&reqData))

			fmt.Print(data, " ")
		case "n2":
			reqData := packet.TNewUserReq{Phone: "12"}
			reqData.Uuid = G_PlayerUuid2
			reqData.Type = packet.NewUser
			reqData.Data = &player.Player{Uuid: reqData.Uuid}

			data, _ := req.Post("http://127.0.0.1:7403/api", req.BodyJSON(&reqData))

			fmt.Print(data, " ")

		case "da1":
			packetToSend := packet.StampPacket(G_PlayerUuid1, nil, packet.DialAddr)

			_, err := packetToSend.SendUdpStream2(udpConnection)
			if err != nil {
				log15.Error("SendUdpStream2", "err", err)
			}
		case "da2":
			packetToSend := packet.StampPacket(G_PlayerUuid2, nil, packet.DialAddr)

			_, err := packetToSend.SendUdpStream2(udpConnection)
			if err != nil {
				log15.Error("SendUdpStream2", "err", err)
			}
		case "p1":

			user.Uuid = G_PlayerUuid1

			packetToSend := packet.StampPacket(G_PlayerUuid1, player.PlayerPosition{X: 1, Y: 12}, packet.UpdatePos)

			_, err := packetToSend.SendUdpStream2(udpConnection)
			if err != nil {
				log15.Error("SendUdpStream2", "err", err)
			}
		case "p2":
			user.Uuid = G_PlayerUuid2

			packetToSend := packet.StampPacket(G_PlayerUuid2, player.PlayerPosition{X: 1, Y: 12}, packet.UpdatePos)

			_, err := packetToSend.SendUdpStream2(udpConnection)
			if err != nil {
				log15.Error("SendUdpStream2", "err", err)
			}

		case "d1":
			reqData := packet.TDisconnectReq{}
			reqData.Uuid = G_PlayerUuid1
			reqData.Type = packet.Disconnect
			data, _ := req.Post("http://127.0.0.1:7403/api", req.BodyJSON(&reqData))
			fmt.Print(data, " ")
		case "d2":
			reqData := packet.TDisconnectReq{}
			reqData.Uuid = G_PlayerUuid2
			reqData.Type = packet.Disconnect
			data, _ := req.Post("http://127.0.0.1:7403/api", req.BodyJSON(&reqData))
			fmt.Print(data, " ")
		case "help", "h":
			log15.Error("help(h)")
			log15.Error("new (n1,n2,n3)")
			log15.Error("dial (da1,da2,da3)")
			log15.Error("pos (p1,p2,p3)")
			log15.Error("disconnet  (d1,d2,d3)")
		default:
			log15.Error("Unknown client command (help,h)", "cmd", command)
			time.Sleep(1 * time.Second)
		}
	}
}
