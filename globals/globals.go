package globals

// used when updating/accessing the player list

var TasksToWin = 0     // total tasks are needed to win
var TasksDone = 0      // how many tasks have been finished
var ImpostorsAlive = 0 // how many impostors are left
var CrewAlive = 0      // how many crew are left
var IsInLobby = true   // whether the game started or not

var QueueReaders int = 5 // amount of UDP queue reader threads

// var LobbyPositions = make([]player.PlayerPosition, 0) // holds where the players spawn when joining the lobby
