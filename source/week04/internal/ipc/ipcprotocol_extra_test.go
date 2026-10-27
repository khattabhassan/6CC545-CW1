package ipc

import (
	"bytes"
	"strings"
	"testing"
)

func TestWeek04_MessageValidationTable(t *testing.T) {
	good := Message{Version: 1, RequestID: "req-1", NodeID: "node-a", Text: "hello"}
	tests := []struct {
		name   string
		mutate func(*Message)
		wantOK bool
	}{
		{"valid", func(m *Message) {}, true},
		{"512 byte text ok", func(m *Message) { m.Text = strings.Repeat("x", 512) }, true},
		{"513 byte text rejected", func(m *Message) { m.Text = strings.Repeat("x", 513) }, false},
		{"empty text rejected", func(m *Message) { m.Text = "" }, false},
		{"invalid utf8 rejected", func(m *Message) { m.Text = "\xff\xfe" }, false},
		{"bad request id rejected", func(m *Message) { m.RequestID = "bad id!" }, false},
		{"empty node id rejected", func(m *Message) { m.NodeID = "" }, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := good
			tc.mutate(&m)
			if err := m.Validate(); (err == nil) != tc.wantOK {
				t.Fatalf("Validate() = %v, wantOK %v", err, tc.wantOK)
			}
		})
	}
}

func TestWeek04_IncompleteFrames(t *testing.T) {
	// header cut short
	if _, err := ReadFrame(bytes.NewReader([]byte{0, 0})); err == nil {
		t.Fatal("incomplete header accepted")
	}
	// header promises 10 bytes, only 3 arrive
	data := append([]byte{0, 0, 0, 10}, []byte("abc")...)
	if _, err := ReadFrame(bytes.NewReader(data)); err == nil {
		t.Fatal("incomplete body accepted")
	}
}