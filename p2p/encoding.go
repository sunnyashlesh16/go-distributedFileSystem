package p2p

import (
	"encoding/gob"
	"fmt"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, message *RPC) error {
	//return gob.NewDecoder(r).Decode(message)
	gob.Register(RPC{})

	// Decode the message
	err := gob.NewDecoder(r).Decode(message)
	if err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}

	return nil
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, message *RPC) error {
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	message.Payload = buf[:n]

	return nil
}
