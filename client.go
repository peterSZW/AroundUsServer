package main

import (
	"aroundUsServer/packet"
	"aroundUsServer/player"
	"fmt"
	"net"
	"strconv"
	"strings"

	"log"
)

func client() {

	var user player.Player
	user.UdpAddress = &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 27403,
	}

	udpConnection, err := net.DialUDP("udp", nil, user.UdpAddress)

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

			_, err = packetToSend.SendUdpStream2(udpConnection)
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
