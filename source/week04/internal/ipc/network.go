package ipc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type OnMessage func(Message)

func deadline(ctx context.Context, timeout time.Duration) time.Time {
	d := time.Now().Add(timeout)
	if cd, ok := ctx.Deadline(); ok && cd.Before(d) {
		return cd
	}
	return d
}
func makeAck(id string, m Message) Ack {
	return Ack{Version: Version, RequestID: m.RequestID, Receiver: id, Accepted: true}
}
func checkAck(a Ack, m Message) error {
	if a.Version != Version || a.RequestID != m.RequestID || !a.Accepted || a.Receiver == "" {
		return errors.New("invalid or unrelated acknowledgement")
	}
	return nil
}
func ServeTCP(ctx context.Context, ln net.Listener, id string, timeout time.Duration, on OnMessage) error {
	// TODO W04_SERVE_TCP: implement the contract described in TASKS.md.
	return ErrUnimplemented
}
func SendTCP(ctx context.Context, address string, m Message, timeout time.Duration) (Ack, error) {
	// TODO W04_SEND_TCP: implement the contract described in TASKS.md.
	return Ack{}, ErrUnimplemented
}
func ServeUDP(ctx context.Context, c *net.UDPConn, id string, timeout time.Duration, on OnMessage) error {
	// TODO W04_SERVE_UDP: implement the contract described in TASKS.md.
	return ErrUnimplemented
}
func SendUDP(ctx context.Context, address string, m Message, timeout time.Duration) (Ack, error) {
	// TODO W04_SEND_UDP: implement the contract described in TASKS.md.
	return Ack{}, ErrUnimplemented
}

var _ = fmt.Errorf
var _ = log.Printf
var _ sync.WaitGroup
var _ = deadline
var _ = makeAck
var _ = checkAck
