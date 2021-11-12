package main

import (
	"fmt"
	"sync"

	uuid "github.com/satori/go.uuid"

	"github.com/prsolucoes/gotslist"
)

func NewUUID() string {

	return uuid.NewV4().String()

}

var Rooms sync.Map

// var Rooms map[string]*gotslist.GoTSList

func GetAllRoomUsers() []string {

	var ids []string
	Rooms.Range(func(k, v interface{}) bool {

		//fmt.Println("iterate:", k, v)
		ids = append(ids, k.(string))
		v.(*sync.Map).Range(func(kk, vv interface{}) bool {
			ids = append(ids, "-"+kk.(string))
			return true
		})

		return true
	})

	// var ids []string
	// for id, _ := range Rooms {
	// 	ids = append(ids, id)
	// }
	return ids

}

func GetRoomIDs() []string {

	var ids []string
	Rooms.Range(func(k, v interface{}) bool {
		// fmt.Println("iterate:", k, v)
		ids = append(ids, k.(string))
		return true
	})

	// var ids []string
	// for id, _ := range Rooms {
	// 	ids = append(ids, id)
	// }
	return ids

}

func GetRoomUsers(id string) *sync.Map {

	v, ok := Rooms.Load(id)
	if ok {
		return v.(*sync.Map)
	} else {
		return nil
	}

}

func JoinRoom(id string, roomID string) bool {
	users := GetRoomUsers(roomID)
	if users == nil {
		return false
	}
	users.Store(id, id)
	return true

}
func JoinNewRoom(id string) string {
	roomid := NewUUID()
	Rooms.Store(roomid, &sync.Map{})
	JoinRoom(id, roomid)
	return roomid
}

func ClearRoom() {

	Rooms.Range(func(k, v interface{}) bool {

		v.(*sync.Map).Range(func(kk, vv interface{}) bool {
			vv.(*sync.Map).Delete(kk)

			return true
		})
		v.(*sync.Map).Delete(k)
		return true
	})

	// for id, tslist := range Rooms {
	// 	for e := tslist.Front(); e != nil; e = e.Next() {
	// 		tslist.Remove(e)
	// 	}
	// 	delete(Rooms, id)

	// }

}

func ExampleHowToUse() {

	type Tclass struct {
		Id   string
		Data string
		Age  int32
	}
	// new
	tslist := gotslist.NewGoTSList()
	// add
	tslist.PushBack("New element")
	// remove
	for e := tslist.Front(); e != nil; e = e.Next() {
		tslist.Remove(e)
	}
	// len
	_ = tslist.Len()
	// is empty
	_ = tslist.IsEmpty()
	// lock and unlock
	tslist.Lock()
	tslist.Unlock()
	fmt.Println("ok")
	// Output: ok
}
