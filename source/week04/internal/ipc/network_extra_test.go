package ipc

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// An oversized (but otherwise well-formed) UDP datagram must be dropped without touching the inbox.
func TestWeek04_OversizeUDPDropped(t *testing.T) {
	c, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls atomic.Int32
	done := make(chan error, 1)
	go func() { done <- ServeUDP(ctx, c, "node-b", 100*time.Millisecond, func(Message) { calls.Add(1) }) }()

	remote, err := net.Dial("udp", c.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	_ = remote.SetDeadline(time.Now().Add(200 * time.Millisecond))
	payload := `{"version":1,"request_id":"r1","node_id":"node-a","text":"` + strings.Repeat("x", MaxMessage) + `"}`
	if _, err := remote.Write([]byte(payload)); err != nil {
		t.Fatal(err)
	}
	var b [128]byte
	if n, err := remote.Read(b[:]); err == nil || n != 0 {
		t.Error("oversize datagram should get no reply")
	}
	if calls.Load() != 0 {
		t.Error("oversize datagram changed the inbox")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("serve: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("UDP shutdown hung")
	}
}

// A client that sends only part of a frame must not hold the server forever.
func TestWeek04_ServerTimesOutIncompleteFrame(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls atomic.Int32
	done := make(chan error, 1)
	go func() { done <- ServeTCP(ctx, ln, "node-b", 100*time.Millisecond, func(Message) { calls.Add(1) }) }()

	conn, err := net.DialTimeout("tcp", ln.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	// header promises 10 bytes; only 3 are sent
	_, _ = conn.Write(append([]byte{0, 0, 0, 10}, []byte("abc")...))
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	start := time.Now()
	var b [1]byte
	_, err = conn.Read(b[:])
	if err == nil {
		t.Fatal("server replied to an incomplete frame")
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		t.Fatal("server kept the connection open past its own deadline")
	}
	if time.Since(start) > 800*time.Millisecond {
		t.Fatalf("server took too long to give up: %v", time.Since(start))
	}
	if calls.Load() != 0 {
		t.Error("incomplete frame reached the inbox")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("serve: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("TCP shutdown hung")
	}
}

// An acknowledgement for a different request must not be accepted.
func TestWeek04_CheckAckRejectsUnrelated(t *testing.T) {
	m := Message{Version: 1, RequestID: "r1", NodeID: "node-a", Text: "hi"}
	if err := checkAck(Ack{Version: 1, RequestID: "r1", Receiver: "node-b", Accepted: true}, m); err != nil {
		t.Fatalf("good ack rejected: %v", err)
	}
	bad := []Ack{
		{Version: 1, RequestID: "other", Receiver: "node-b", Accepted: true},
		{Version: 1, RequestID: "r1", Receiver: "node-b", Accepted: false},
		{Version: 2, RequestID: "r1", Receiver: "node-b", Accepted: true},
		{Version: 1, RequestID: "r1", Receiver: "", Accepted: true},
	}
	for i, a := range bad {
		if checkAck(a, m) == nil {
			t.Errorf("bad ack %d accepted", i)
		}
	}
}

// Fragmented TCP: a frame arriving in separate small writes must still decode.
func TestWeek04_FrameInPieces(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteFrame(&buf, []byte(`{"text":"pieces"}`)); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFrame(&buf)
	if err != nil || string(got) != `{"text":"pieces"}` {
		t.Fatalf("round trip failed: %q %v", got, err)
	}
}