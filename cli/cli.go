package cli

import (
	"aroundUsServer/player"
	"fmt"
	"log"
	"strconv"
)

func ServerConsoleCLI() {
	for {
		// var command string
		// fmt.Scanln(&command)
		var command, parameter string
		fmt.Scanln(&command, &parameter)

		//commands := strings.Split(strings.Trim(command, "\n\t /\\'\""), " ")
		switch command {
		case "help", "h":
			log.Println("help(h)")
			log.Println("list(ls)")
			log.Println("disconnet(dc) [id]")
		case "list", "ls":
			for _, player1 := range player.PlayerList {
				player1.PrintUser()
			}
		case "disconnet", "dc":
			_, err := strconv.Atoi(parameter)
			if err != nil {
				log.Println("Cant convert to number position")
			}
			// globals.PlayerList[id].TcpConnection.Close()
		default:
			log.Println("Unknown command")
		}
	}
}
