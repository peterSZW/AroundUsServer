package packet

import (
	"aroundUsServer/player"
	"encoding/json"
	"fmt"
	"net"
)

const (
	NewUser                = iota + 1 // TCP// Client -> Server packets
	GetUser                           // TCP
	Auth                              // TCP
	Disconnect                        // TCP
	GetRooms                          // TCP
	GetRoomUsers                      // TCP
	JoinRoom                          // TCP
	JoinNewRoom                       // TCP
	LeaveRoom                         // TCP
	Error                             // tcp
	GL_UsersIn                        // TCP// Server -> Client packets
	GL_IsUserManager                  // TCP
	GL_NewPlayerConnected             // TCP
	GL_ClientSpawnPosition            // TCP
	GL_GameOver                       // TCP
	GL_PlayerDied                     // TCP
	GL_KilledPlayer                   // TCP
	GL_Init                           // TCP
	GL_StartGame                      // TCP
)

type TBaseReqPacket struct {
	Type  int16  `json:"type"`
	Seq   int64  `json:"seq"`
	Uuid  string `json:"uuid"`
	Token string `json:"token"`
}

type TBaseRspPacket struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	MsgEx string `json:"msgex"`
}

//====

type TNewUserReq struct {
	TBaseReqPacket
	Phone string         `json:"phone"`
	Email string         `json:"email"`
	Data  *player.Player `json:"data"`
}
type TNewUserRsp struct {
	TBaseReqPacket
	TBaseRspPacket
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type TAuthReq struct {
	TBaseReqPacket
	Phone string `json:"phone"`
	Email string `json:"email"`
	Pass  string `json:"pass"`
}
type TAuthRsp struct {
	TBaseReqPacket
	TBaseRspPacket
}

type TDisconnectReq struct {
	TBaseReqPacket
}
type TDisconnectRsp struct {
	TBaseRspPacket
}

//===
type TGetRoomsReq struct {
	TBaseReqPacket
}
type TGetRoomsRsp struct {
	TBaseRspPacket
}

//===
type TGetRoomUsersReq struct {
	TBaseReqPacket
}
type TGetRoomUsersRsp struct {
	TBaseRspPacket
}

type TJoinRoomReq struct {
	TBaseReqPacket
}
type TJoinRoomRsp struct {
	TBaseRspPacket
}

type TJoinNewRoomReq struct {
	TBaseReqPacket
}
type TJoinNewRoomRsp struct {
	TBaseRspPacket
}

type TLeaveRoomReq struct {
	TBaseReqPacket
}
type TLeaveRoomRsp struct {
	TBaseRspPacket
}

type TDialAddrReq struct {
	TBaseReqPacket
}
type TDialAddrRsp struct {
	TBaseRspPacket
}

type TUpdatePosReq struct {
	TBaseReqPacket
	PP player.PlayerPosition
}
type TUpdatePosRsp struct {
	TBaseRspPacket
}

type TUpdateRotationReq struct {
	TBaseReqPacket
	PP player.PlayerRotation
}
type TUpdateRotationRsp struct {
	TBaseRspPacket
}

type TPositionBroadcastReq struct {
	TBaseReqPacket
	PP player.PlayerPosition
}
type TPositionBroadcastRsp struct {
	TBaseRspPacket
}

type THeartBeatReq struct {
	TBaseReqPacket
}
type THeartBeatRsp struct {
	TBaseRspPacket
}

const (
	DialAddr          = iota + 1000 // UDP
	UpdatePos                       // UDP
	UpdateRotation                  // UDP
	PositionBroadcast               // UDP
	HeartBeat                       // UDP
)

type ClientPacketRaw struct {
	Type int16  `json:"type"`
	Seq  int64  `json:"seq"`
	Uuid string `json:"uuid"`
	//Data interface{} `json:"data"`
}

type ClientPacket struct {
	Uuid string                 `json:"uuid"`
	Type int16                  `json:"type"`
	Seq  int64                  `json:"seq"`
	Data map[string]interface{} `json:"data"`
}

type ServerPacket struct {
	Type int16       `json:"type"`
	Seq  int64       `json:"seq"`
	Uuid string      `json:"uuid"`
	Data interface{} `json:"data"`
}

type GameInitData struct {
	Imposters    []string `json:"imposters"`
	TaskCount    uint16   `json:"taskCount"`
	PlayerSpeed  uint16   `json:"playerSpeed"`
	KillCooldown uint16   `json:"killCooldown"`
	Emergencies  uint16   `json:"emergencies"`
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

func StampPacket(uuid string, data interface{}, packetType int16) ServerPacket {
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
