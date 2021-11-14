package cli

import (
	"aroundUsServer/player"
	"fmt"
	"strconv"
	"time"

	"github.com/inconshreveable/log15"
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

			player.PlayerMap.Range(func(k, v interface{}) bool {
				user := v.(*player.Player)
				user.PrintUser()

				return true
			})

		case "disconnet", "dc":
			_, err := strconv.Atoi(parameter)
			if err != nil {
				log15.Error("Cant convert to number position")
			}

		default:
			log15.Error("Unknown command1", "cmd", command)
			time.Sleep(1 * time.Second)
		}
	}
}
