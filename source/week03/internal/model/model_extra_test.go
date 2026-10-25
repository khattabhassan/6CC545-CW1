package model

import "testing"

func TestWeek03_PeerEdgeCases(t *testing.T) {
	good := Peer{ID: "node-b", HTTPAddress: "127.0.0.1:8082", GRPCAddress: "127.0.0.1:9092", Status: "unknown"}
	tests := []struct {
		name   string
		mutate func(*Peer)
		wantOK bool
	}{
		{"valid", func(p *Peer) {}, true},
		{"ipv6 loopback ok", func(p *Peer) { p.HTTPAddress = "[::1]:8082" }, true},
		{"status online ok", func(p *Peer) { p.Status = "online" }, true},
		{"malformed status", func(p *Peer) { p.Status = "maybe" }, false},
		{"non-loopback rejected", func(p *Peer) { p.HTTPAddress = "8.8.8.8:80" }, false},
		{"bad id", func(p *Peer) { p.ID = "bad id!" }, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := good
			tc.mutate(&p)
			if err := p.Validate(); (err == nil) != tc.wantOK {
				t.Fatalf("Validate() = %v, wantOK %v", err, tc.wantOK)
			}
		})
	}
}