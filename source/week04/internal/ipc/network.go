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

const maxHandlers = 32 // cap on simultaneous TCP connection handlers

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

// ---------- TCP ----------

// ServeTCP accepts connections until ctx is cancelled. A bad connection is closed;
// the listener keeps running. It does not return until all handlers have finished.
func ServeTCP(ctx context.Context, ln net.Listener, id string, timeout time.Duration, on OnMessage) error {
	var wg sync.WaitGroup
	defer wg.Wait() // runs last: never return while handlers are still running

	// Cancellation must interrupt a blocked Accept, so close the listener.
	stopClose := context.AfterFunc(ctx, func() { _ = ln.Close() })
	defer stopClose()

	sem := make(chan struct{}, maxHandlers)
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil // normal stop
			}
			return fmt.Errorf("tcp accept: %w", err)
		}
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			_ = conn.Close()
			return nil
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			handleTCP(ctx, conn, id, timeout, on)
		}()
	}
}

// handleTCP serves ONE request on ONE connection. All errors close only this connection.
func handleTCP(ctx context.Context, conn net.Conn, id string, timeout time.Duration, on OnMessage) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout)) // bounded lifetime
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	data, err := ReadFrame(conn)
	if err != nil {
		log.Printf("tcp %s: rejected frame: %v", conn.RemoteAddr(), err)
		return
	}
	var m Message
	if err := Decode(data, &m); err != nil {
		log.Printf("tcp %s: rejected message: %v", conn.RemoteAddr(), err)
		return
	}
	if err := m.Validate(); err != nil {
		log.Printf("tcp %s: invalid message: %v", conn.RemoteAddr(), err)
		return
	}
	if on != nil {
		on(m) // only valid messages reach the inbox
	}
	reply, err := Encode(makeAck(id, m))
	if err != nil {
		log.Printf("tcp: encode ack: %v", err)
		return
	}
	if err := WriteFrame(conn, reply); err != nil {
		log.Printf("tcp %s: write ack: %v", conn.RemoteAddr(), err)
	}
}

// SendTCP performs one bounded request/ack round trip.
func SendTCP(ctx context.Context, address string, m Message, timeout time.Duration) (Ack, error) {
	if err := m.Validate(); err != nil {
		return Ack{}, fmt.Errorf("outgoing message invalid: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return Ack{}, err
	}
	dl := deadline(ctx, timeout) // ONE overall budget
	opCtx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(opCtx, "tcp", address)
	if err != nil {
		return Ack{}, fmt.Errorf("tcp dial: %w", err)
	}
	defer conn.Close() // closed on EVERY return path
	_ = conn.SetDeadline(dl)
	stop := context.AfterFunc(opCtx, func() { _ = conn.Close() }) // cancellation interrupts blocked I/O
	defer stop()

	data, err := Encode(m)
	if err != nil {
		return Ack{}, err
	}
	if err := WriteFrame(conn, data); err != nil {
		return Ack{}, fmt.Errorf("tcp send: %w", err)
	}
	reply, err := ReadFrame(conn)
	if err != nil {
		return Ack{}, fmt.Errorf("tcp receive: %w", err)
	}
	var a Ack
	if err := Decode(reply, &a); err != nil {
		return Ack{}, fmt.Errorf("tcp ack decode: %w", err)
	}
	if err := checkAck(a, m); err != nil {
		return Ack{}, err
	}
	return a, nil
}

// ---------- UDP ----------

// ServeUDP reads datagrams until ctx is cancelled. Invalid datagrams are dropped silently
// (logged), so a client sees a timeout rather than an error reply.
func ServeUDP(ctx context.Context, c *net.UDPConn, id string, timeout time.Duration, on OnMessage) error {
	// Cancellation must interrupt a blocked read, so close the socket.
	stop := context.AfterFunc(ctx, func() { _ = c.Close() })
	defer stop()

	buf := make([]byte, MaxMessage+1) // one byte larger than the limit reveals oversize datagrams
	for {
		n, addr, err := c.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil // normal stop
			}
			if errors.Is(err, net.ErrClosed) {
				return err
			}
			log.Printf("udp read error (continuing): %v", err)
			continue
		}
		if n > MaxMessage {
			log.Printf("udp %s: dropped oversize datagram", addr)
			continue
		}
		var m Message
		if err := Decode(buf[:n], &m); err != nil {
			log.Printf("udp %s: dropped malformed datagram: %v", addr, err)
			continue
		}
		if err := m.Validate(); err != nil {
			log.Printf("udp %s: dropped invalid message: %v", addr, err)
			continue
		}
		if on != nil {
			on(m)
		}
		reply, err := Encode(makeAck(id, m))
		if err != nil {
			log.Printf("udp: encode ack: %v", err)
			continue
		}
		_ = c.SetWriteDeadline(time.Now().Add(timeout))
		if _, err := c.WriteToUDP(reply, addr); err != nil {
			log.Printf("udp %s: write ack: %v", addr, err)
		}
	}
}

// SendUDP sends one datagram and waits for one acknowledgement within the budget. No retries.
func SendUDP(ctx context.Context, address string, m Message, timeout time.Duration) (Ack, error) {
	if err := m.Validate(); err != nil {
		return Ack{}, fmt.Errorf("outgoing message invalid: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return Ack{}, err
	}
	dl := deadline(ctx, timeout)
	opCtx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(opCtx, "udp", address)
	if err != nil {
		return Ack{}, fmt.Errorf("udp dial: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(dl)
	stop := context.AfterFunc(opCtx, func() { _ = conn.Close() })
	defer stop()

	data, err := Encode(m) // no length prefix: the datagram IS the message
	if err != nil {
		return Ack{}, err
	}
	if _, err := conn.Write(data); err != nil {
		return Ack{}, fmt.Errorf("udp send: %w", err)
	}
	buf := make([]byte, MaxMessage+1)
	n, err := conn.Read(buf)
	if err != nil {
		return Ack{}, fmt.Errorf("udp receive (no ack before deadline?): %w", err)
	}
	if n > MaxMessage {
		return Ack{}, errors.New("oversize acknowledgement datagram")
	}
	var a Ack
	if err := Decode(buf[:n], &a); err != nil {
		return Ack{}, fmt.Errorf("udp ack decode: %w", err)
	}
	if err := checkAck(a, m); err != nil {
		return Ack{}, err
	}
	return a, nil
}