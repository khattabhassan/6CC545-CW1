package config

import (
	"io"
	"testing"
	"time"
)

func TestSmoke_ParseHelp(t *testing.T) {
	if _, e := Parse([]string{"-help"}, io.Discard); e == nil {
		t.Fatal("help must return flag.ErrHelp")
	}
}
func TestWeek01_Config(t *testing.T) {
	good := Config{ID: "node-a", HTTP: "127.0.0.1:8081", Timeout: 2 * time.Second}
	if e := good.Validate(); e != nil {
		t.Fatalf("valid configuration rejected: %v", e)
	}
	cases := []Config{good, good, good, good, good}
	cases[0].ID = ""
	cases[1].HTTP = ":8081"
	cases[2].HTTP = "127.0.0.1:0"
	cases[3].Timeout = 0
	cases[4].HTTP = "0.0.0.0:8081"
	for i, c := range cases {
		if c.Validate() == nil {
			t.Errorf("invalid case %d accepted", i)
		}
	}
}
func TestWeek01_UnknownFlag(t *testing.T) {
	if _, e := Parse([]string{"-bogus"}, io.Discard); e == nil {
		t.Fatal("unknown flag accepted")
	}
}
