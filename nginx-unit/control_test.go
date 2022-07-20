package unit

import (
	"encoding/json"
	"testing"
)

func jsonString(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

var config interface{} = nil

func TestGet(t *testing.T) {
	control := NewControl("/usr/local/var/run/unit/control.sock")
	rs, err := control.Get("http://localhost/config")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(jsonString(rs))
	config = rs
}

func TestPut(t *testing.T) {
	control := NewControl("/usr/local/var/run/unit/control.sock")
	err := control.Put("http://localhost/config", config)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("SUCCESS")
}
