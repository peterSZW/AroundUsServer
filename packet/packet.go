package packet

import (
	"encoding/json"
	"fmt"
	"net"
)

const (
	InitUser                      = iota + 1 // TCP// Client -> Server packets
	DialAddr                                 // UDP
	UpdatePos                                // UDP
	UpdateRotation                           // UDP
	PositionBroadcast                        // UDP
	HeartBeat                                // UDP
	UserDisconnected                         // TCP
	UserId                                   // TCP
	Error                                    // TCP
	GameLogic_UsersInGame                    // TCP// Server -> Client packets
	GameLogic_IsUserManager                  // TCP
	GameLogic_NewPlayerConnected             // TCP
	GameLogic_ClientSpawnPosition            // TCP
	GameLogic_GameOver                       // TCP
	GameLogic_PlayerDied                     // TCP
	GameLogic_KilledPlayer                   // TCP
	GameLogic_Init                           // TCP
	GameLogic_StartGame                      // TCP
)

type ClientPacketRaw struct {
	Type int8   `json:"type"`
	Seq  int64  `json:"seq"`
	Uuid string `json:"uuid"`
	//Data interface{} `json:"data"`
}

type ClientPacket struct {
	Uuid string                 `json:"uuid"`
	Type int8                   `json:"type"`
	Seq  int64                  `json:"seq"`
	Data map[string]interface{} `json:"data"`
}

type ServerPacket struct {
	Type int8        `json:"type"`
	Seq  int64       `json:"seq"`
	Uuid string      `json:"uuid"`
	Data interface{} `json:"data"`
}

type GameInitData struct {
	Imposters    []string `json:"imposters"`
	TaskCount    uint8    `json:"taskCount"`
	PlayerSpeed  uint8    `json:"playerSpeed"`
	KillCooldown uint8    `json:"killCooldown"`
	Emergencies  uint8    `json:"emergencies"`
}

func (dataPacket *ClientPacket) DataToBytes() ([]byte, error) {
	// buf, err := helpers.GetBytes(dataPacket.Data)
	jsonString, err := json.Marshal(dataPacket.Data)
	if err != nil {
		return nil, fmt.Errorf("error while turning packet data to bytes")
	}
	return []byte(jsonString), nil
}

var seq int64

func StampPacket(uuid string, data interface{}, packetType int8) ServerPacket {
	seq++
	return ServerPacket{Uuid: uuid, Type: packetType, Seq: seq, Data: data}
	//ServerPacket{}
}

func (packet *ServerPacket) SendTcpStream(tcpConnection net.Conn) (int, error) {
	packetJson, err := json.Marshal(*packet)
	if err != nil {
		return 0, fmt.Errorf("error while marshaling TCP packet")
	}
	return tcpConnection.Write([]byte(packetJson))
}

func (packet *ServerPacket) SendUdpStream(udpConnection *net.UDPConn, udpAddress *net.UDPAddr) (int, error) {
	packetJson, err := json.Marshal(*packet)
	if err != nil {
		return 0, fmt.Errorf("error while marshaling UDP packet")
	}
	return udpConnection.WriteToUDP([]byte(packetJson), udpAddress)
}
func (packet *ServerPacket) SendUdpStream2(udpConnection *net.UDPConn) (int, error) {
	packetJson, err := json.Marshal(*packet)
	if err != nil {
		return 0, fmt.Errorf("error while marshaling UDP packet")
	}
	return udpConnection.Write([]byte(packetJson))
}
