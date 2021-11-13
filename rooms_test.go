package main

import (
	"aroundUsServer/packet"
	"fmt"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/imroc/req"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRooms(t *testing.T) {
	roomid := JoinNewRoom("ABC")
	JoinRoom("FFFF", roomid)
	JoinNewRoom("CDE")
	fmt.Println(GetRoomIDs())
	// fmt.Println(Rooms)
	fmt.Println(GetAllRoomUsers())
}

func TestSorm(t *testing.T) {

	reqData := packet.TAuthReq{Phone: "12"}
	reqData.Type = packet.Auth

	Convey("测试:", t, func() {
		data, err := req.Post("http://127.0.0.1:7403/api", req.BodyJSON(&reqData))

		fmt.Print(data, " ")

		js, _ := simplejson.NewJson(data.Bytes())

		So(err == nil, ShouldBeTrue)
		So(js.Get("code").MustInt() == 500, ShouldBeTrue)

	})

}
