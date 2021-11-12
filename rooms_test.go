package main

import (
	"fmt"
	"testing"
)

func TestRooms(t *testing.T) {
	roomid := JoinNewRoom("ABC")
	JoinRoom("FFFF", roomid)
	JoinNewRoom("CDE")
	fmt.Println(GetRoomIDs())
	// fmt.Println(Rooms)
	fmt.Println(GetAllRoomUsers())
}
