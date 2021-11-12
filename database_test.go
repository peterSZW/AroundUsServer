package main

import (
	"fmt"
	"testing"
)

func TestSAVEUUID(t *testing.T) {
	logic_save_uuid("ABC")
	fmt.Println(logic_get_uuid("ABC"))
	fmt.Println(logic_get_uuids())
}
