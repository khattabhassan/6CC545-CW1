package model

import (
	"encoding/hex"
	"errors"
	"example.edu/6cc545/cw1/internal/config"
	"strings"
)

const maxFileName = 128

type Peer struct {
	ID          string `json:"id"`
	HTTPAddress string `json:"http_address"`
	GRPCAddress string `json:"grpc_address"`
	Status      string `json:"status"`
}
type FileMetadata struct {
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	SHA256       string `json:"sha256"`
	ModifiedUnix int64  `json:"modified_unix"`
}

// Validate checks syntax only. It never contacts the peer.
func (p Peer) Validate() error {
	if !config.ValidID(p.ID) {
		return errors.New("invalid peer id: use 1-32 letters, digits, '_' or '-'")
	}
	if !config.ValidAddress(p.HTTPAddress) {
		return errors.New("invalid http_address: need a loopback host:port")
	}
	if !config.ValidAddress(p.GRPCAddress) {
		return errors.New("invalid grpc_address: need a loopback host:port")
	}
	switch p.Status {
	case "unknown", "online", "offline":
	default:
		return errors.New("invalid status: must be unknown, online or offline")
	}
	return nil
}

func (f FileMetadata) Validate() error {
	if f.Name == "" || len(f.Name) > maxFileName {
		return errors.New("invalid file name: must be 1-128 bytes")
	}
	if f.Name == "." || f.Name == ".." || strings.ContainsAny(f.Name, "/\\\x00") {
		return errors.New("invalid file name: no directory separators, null bytes, '.' or '..'")
	}
	if f.Size < 0 {
		return errors.New("invalid size: must not be negative")
	}
	if f.ModifiedUnix < 0 {
		return errors.New("invalid modified_unix: must not be negative")
	}
	if len(f.SHA256) != 64 {
		return errors.New("invalid sha256: need exactly 64 hex characters")
	}
	if _, err := hex.DecodeString(f.SHA256); err != nil {
		return errors.New("invalid sha256: not hexadecimal")
	}
	return nil
}