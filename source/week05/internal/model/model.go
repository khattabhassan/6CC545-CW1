package model

import (
	"encoding/hex"
	"errors"
	"example.edu/6cc545/cw1/internal/config"
	"strings"
)

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

func (p Peer) Validate() error {
	// TODO W03_PEER: implement the contract described in TASKS.md.
	return errors.New("TODO W03_PEER")
}
func (f FileMetadata) Validate() error {
	// TODO W03_FILE: implement the contract described in TASKS.md.
	return errors.New("TODO W03_FILE")
}

var _ = hex.DecodeString
var _ = strings.ContainsAny
var _ = config.ValidID
