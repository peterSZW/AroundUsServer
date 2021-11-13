package cli

import (
	"aroundUsServer/player"
	"fmt"
	"strconv"

	"github.com/xiaomi-tc/log15"
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
			log15.Error("help(h)")
			log15.Error("list(ls)")
			log15.Error("disconnet(dc) [id]")
		case "list", "ls":
			for _, player1 := range player.PlayerList {
				player1.PrintUser()
			}
		case "disconnet", "dc":
			_, err := strconv.Atoi(parameter)
			if err != nil {
				log15.Error("Cant convert to number position")
			}
			// globals.PlayerList[id].TcpConnection.Close()
		default:
			log15.Error("Unknown command")
		}
	}
}
