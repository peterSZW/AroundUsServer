package player

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/inconshreveable/log15"
)

var SpawnPositionsStack = make([]PlayerPosition, 100) // holds where the players spawn when respawning after a meeting, functions as a stack

var PlayerList = make(map[string]*Player, 10) // holds the players, maximum 10
var PlayerListLock sync.Mutex
var Colors [12]bool // holds the colors, index indicated the color and the value if its taken or not
var CurrId int      // the next player id when joining
type Player struct {
	Uuid           string         `json:"uuid"`           // Id of the player
	Name           string         `json:"name"`           // The name of the player, can contain anything
	Color          int16          `json:"color"`          // The index of the color in the color list held in the client
	IsManager      bool           `json:"-"`              // Whether the player is the game manager or not, he can start the game
	IsImposter     bool           `json:"isImposter"`     // Sent on the round start to tell the client if hes an imposter or crew
	InVent         bool           `json:"inVent"`         // If true the server shouldnt send the player locations until hes leaving the vent
	IsDead         bool           `json:"isDead"`         // If the player is dead the server shouldnt send his location
	GotReported    bool           `json:"gotReported"`    // If the player didnt get reported yet tell the client to show a body on the ground
	PlayerPosition PlayerPosition `json:"playerPosition"` // The position of the player in Unity world cordinates
	Rotation       PlayerRotation `json:"rotation"`       // Pitch: -90 <= pitch <= 90(head up and down), Yaw: 0 <= rotation <= 360(body rotation)
	TcpConnection  net.Conn       `json:"-"`              // The player TCP connection socket
	UdpAddress     *net.UDPAddr   `json:"-"`              // The player UDP address socket
}

type PlayerPosition struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

type PlayerRotation struct {
	Pitch float32 `json:"rotation"`
	Yaw   float32 `json:"yaw"`
}

func (newPlayer *Player) InitializePlayer() *Player {

	log15.Error("====init=======", newPlayer)

	//newPlayer.TcpConnection = tcpConnection // Set the player TCP connection socket

	// check if the name is taken or invalid
	// we need to keep a counter so the name will be in the format `<name> <count>`

	var newNameCount int16
	var nameOk bool
	oldName := newPlayer.Name

	for !nameOk {
		nameOk = true
		for _, player := range PlayerList {
			if player.Name == newPlayer.Name {
				newNameCount++
				nameOk = false
				newPlayer.Name = fmt.Sprintf("%s %d", oldName, newNameCount)
				break
			}
		}
	}

	if newNameCount == 0 {
		newPlayer.Name = oldName
	}

	// check if the color is taken or invalid, if it is assign next not taken color
	if int16(0) > newPlayer.Color || int16(len(Colors)) <= newPlayer.Color || Colors[newPlayer.Color] {
		for index, color := range Colors {
			if !color {
				newPlayer.Color = int16(index)
				break
			}
		}
	}

	Colors[newPlayer.Color] = true // set player color as taken

	// check if he is the first one in the lobby, if true set the player to be the game manager
	if len(PlayerList) == 0 {
		newPlayer.IsManager = true
	}

	// set player ID and increase to next one, theoretically this can roll back at 2^31-1

	//newPlayer.Uuid = strconv.Itoa(CurrId)
	//CurrId++

	// set player spawn position
	newPlayer.PlayerPosition = SpawnPositionsStack[len(SpawnPositionsStack)-1] // peek at the last element
	SpawnPositionsStack = SpawnPositionsStack[:len(SpawnPositionsStack)]       // pop

	log15.Error("New player got generated:")
	newPlayer.PrintUser()

	return newPlayer
}

func (playerToDelete *Player) DeInitializePlayer() error {
	PlayerListLock.Lock()
	defer PlayerListLock.Unlock()

	delete(PlayerList, playerToDelete.Uuid)

	// give another player the manager
	if playerToDelete.IsManager {
		for _, nextPlayer := range PlayerList {
			nextPlayer.IsManager = true
			break
		}
	}

	// free the color
	Colors[playerToDelete.Color] = false

	playerToDelete = nil

	return nil
}

func (user *Player) PrintUser() {
	//p, err := json.MarshalIndent(user, "", " ")
	p, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}
