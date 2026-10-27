package ipc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"example.edu/6cc545/cw1/internal/config"
	"io"
	"unicode/utf8"
)

const MaxMessage = 4096
const Version = 1

var ErrUnimplemented = errors.New("TODO Week 4: implement IPC")

type Message struct {
	Version   int    `json:"version"`
	RequestID string `json:"request_id"`
	NodeID    string `json:"node_id"`
	Text      string `json:"text"`
}
type Ack struct {
	Version   int    `json:"version"`
	RequestID string `json:"request_id"`
	Receiver  string `json:"receiver"`
	Accepted  bool   `json:"accepted"`
}

func (m Message) Validate() error {
	// TODO W04_VALIDATE: implement the contract described in TASKS.md.
	return ErrUnimplemented
}
func Encode(v any) ([]byte, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	if len(b) == 0 || len(b) > MaxMessage {
		return nil, errors.New("message exceeds limit")
	}
	return b, nil
}
func Decode(data []byte, v any) error {
	// TODO W04_DECODE: implement the contract described in TASKS.md.
	return ErrUnimplemented
}
func WriteFrame(w io.Writer, data []byte) error {
	// TODO W04_WRITE_FRAME: implement the contract described in TASKS.md.
	return ErrUnimplemented
}
func ReadFrame(r io.Reader) ([]byte, error) {
	// TODO W04_READ_FRAME: implement the contract described in TASKS.md.
	return nil, ErrUnimplemented
}

var _ = binary.BigEndian
var _ = bytes.NewReader
var _ = utf8.ValidString
var _ = config.ValidID
