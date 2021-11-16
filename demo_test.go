package main

import (
	"fmt"
	"syscall"
	"time"

	"github.com/gavv/httpexpect"
	. "github.com/smartystreets/goconvey/convey"

	"bytes"

	"github.com/bitly/go-simplejson"

	//log "github.com/inconshreveable/log15"
	"net/http"
	"testing"
)

var testurl string = "http://127.0.0.1/"

// var db *gorm.DB

// func init() {
// 	MyUser := "gw_slc"
// 	Password := "Sup_Lcgw07232020"
// 	Host := "172.16.1.250"
// 	Port := 33306
// 	Db := "supplierchain"

// 	connArgs := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", MyUser, Password, Host, Port, Db)
// 	db, _ = gorm.Open("mysql", connArgs)

// }
func StringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestGoConvey1(t *testing.T) {

	Convey("需要 TestStringSliceEqual  返回 True 当 a != nil  && b != nil", t, func() {
		a := []string{"hello", "goconvey"}
		b := []string{"hello", "goconvey"}
		So(StringSliceEqual(a, b), ShouldBeTrue)
	})
}

func TestTime(t *testing.T) {
	sec, Usec, _ := TimeUsec()
	fmt.Println(sec, Usec)
	time.Sleep(100 * time.Millisecond)
	sec, Usec, _ = TimeUsec()
	fmt.Println(sec, Usec)
}


//func TestGoConvey2(t *testing.T) {

// Convey("调用 ActivityCurrentUser 要返回活动", t, func() {
// 	json := map[string]interface{}{
// 		"data":     map[string]interface{}{"user_id": "10714a9a-a857-4923-a40f-d3654046af8a"},
// 		"usertype": 33333,
// 	}

// 	e := httpexpect.New(t, testurl) //创建一个httpexpect实例
// 	obj := e.POST("/HMActivity/ActivityCurrentUser").
// 		WithJSON(json).
// 		Expect().
// 		Status(http.StatusOK). //判断请求是否200
// 		JSON().
// 		Object()

// 	obj.ContainsKey("code").ValueEqual("code", 0)

// 	iid := obj.Value("data").Raw().(map[string]interface{})["activity"].(map[string]interface{})["iid"].(float64)

// 	So(iid > 0, ShouldBeTrue)
// })
//}

func TestHttpExpect(t *testing.T) {

	Convey("调用 TestHttpExpect  ", t, func() {

		json := map[string]interface{}{
			"data":     map[string]interface{}{"user_id": "10714a9a-a857-4923-a40f-d3654046af8a"},
			"usertype": 33333,
		}
		e := httpexpect.New(t, testurl) //创建一个httpexpect实例
		obj := e.POST("/hotmall/SupplierChain/CategoryList").
			WithJSON(json).
			Expect().
			Status(http.StatusOK). //判断请求是否200
			JSON().
			Object()
		obj.ContainsKey("code").ValueEqual("code", 0)

	})

}

func TestSimpleJson(t *testing.T) {
	Convey("调用 simplejson 能够读取数据", t, func() {
		buf := bytes.NewBuffer([]byte(`{
			"test": {
				"array": [1, "2", 3],
				"arraywithsubs": [
					{"subkeyone": 1},
					{"subkeytwo": 2, "subkeythree": 3}
				],
				"bignum": 9223372036854775807,
				"uint64": 18446744073709551615
			}
		}`))

		js, _ := simplejson.NewFromReader(buf)

		//fmt.Println(js.Get("test").Get("array").Array())

		So(js.Get("test").Get("array").GetIndex(0).MustInt64() == 1, ShouldBeTrue)
	})

}
