package main

import (
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"log"
)

func getIncomingClientUdp(udpConnection *net.UDPConn) {
	err := error(nil)
	fmt.Println("Client listen....")

	for err == nil {
		buffer := make([]byte, 1024)

		size, addr, err := udpConnection.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Cant read packet!", err)
			continue
		}
		log.Println(addr)
		data := buffer[:size]

		var dataPacket packet.ClientPacket
		err = json.Unmarshal(data, &dataPacket)
		if err != nil {
			log.Println("Couldn't parse json player data! Skipping iteration!")
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

func ClientConsoleCLI(udpConnection *net.UDPConn) {

	for {
		var command string
		fmt.Scanln(&command)
		commands := strings.Split(strings.Trim(command, "\n\t /\\'\""), " ")
		switch commands[0] {
		case "help", "h":
			log.Println("help(h)")
			log.Println("login(lg)")
			log.Println("disconnet(dc) [id]")
		case "login", "lg":
			packetToSend := packet.StampPacket(user, packet.DialAddr)

			_, err := packetToSend.SendUdpStream2(udpConnection)
			if err != nil {
				log.Println(err)
			}
		case "init", "it", "1":
			user.Name = "peter"
			user.Color = 1
			user.Id = 1
			packetToSend := packet.StampPacket(user, packet.InitUser)

			_, err := packetToSend.SendUdpStream2(udpConnection)
			if err != nil {
				log.Println(err)
			}
		case "2":
			user.Name = "leo"
			user.Color = 2
			user.Id = 2
			packetToSend := packet.StampPacket(user, packet.InitUser)

			_, err := packetToSend.SendUdpStream2(udpConnection)
			if err != nil {
				log.Println(err)
			}
		case "3":
			user.Name = "alex"
			user.Color = 3
			user.Id = 3
			packetToSend := packet.StampPacket(user, packet.InitUser)

			_, err := packetToSend.SendUdpStream2(udpConnection)
			if err != nil {
				log.Println(err)
			}
		case "disconnet", "dc":
			_, err := strconv.Atoi(commands[1])
			if err != nil {
				log.Println("Cant convert to number position")
			}
		default:
			log.Println("Unknown command")
		}
	}
}
