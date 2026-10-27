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
const maxText = 512

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

// Validate checks meaning (decoding checks only shape).
func (m Message) Validate() error {
	if m.Version != Version {
		return errors.New("unsupported version")
	}
	if !config.ValidID(m.RequestID) {
		return errors.New("invalid request_id")
	}
	if !config.ValidID(m.NodeID) {
		return errors.New("invalid node_id")
	}
	if len(m.Text) < 1 || len(m.Text) > maxText {
		return errors.New("text must be 1-512 bytes")
	}
	if !utf8.ValidString(m.Text) {
		return errors.New("text is not valid UTF-8")
	}
	return nil
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

// Decode accepts exactly one bounded JSON object with no unknown fields.
func Decode(data []byte, v any) error {
	if len(data) == 0 || len(data) > MaxMessage {
		return errors.New("message size out of range")
	}
	if t := bytes.TrimSpace(data); len(t) == 0 || t[0] != '{' {
		return errors.New("expected a JSON object")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return errors.New("unexpected data after the JSON object")
	}
	return nil
}

// WriteFrame writes a 4-byte big-endian length, then the data, handling short writes.
func WriteFrame(w io.Writer, data []byte) error {
	if len(data) == 0 || len(data) > MaxMessage {
		return errors.New("frame size out of range")
	}
	buf := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(data)))
	copy(buf[4:], data)
	for len(buf) > 0 {
		n, err := w.Write(buf)
		if err != nil {
			return err
		}
		if n <= 0 {
			return io.ErrShortWrite
		}
		buf = buf[n:]
	}
	return nil
}

// ReadFrame reads a full 4-byte header, checks the length, THEN allocates the body.
func ReadFrame(r io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(header[:])
	if n == 0 || n > MaxMessage {
		return nil, errors.New("frame length out of range")
	}
	body := make([]byte, n)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}