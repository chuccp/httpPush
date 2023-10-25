package util

import (
	"encoding/json"
	"testing"
)

type Page struct {
	Num int
}

func TestNamePtr(t *testing.T) {

	var local any
	local = &Page{Num: 0}

	k := NewPtr(local)

	err := json.Unmarshal([]byte(`{"Num":1}`), k)
	if err != nil {
		return
	}

	v := k.(*interface{})

	t.Log(*v)

}
