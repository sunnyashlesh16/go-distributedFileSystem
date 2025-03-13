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
	fmt.Println("Decoding the Message")
	peerbuf := make([]byte, 1)
	_, err := r.Read(peerbuf)
	if err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}
	fmt.Printf("Decoding a the Byte%v\n", peerbuf[0])

	if peerbuf[0] == IncomingStream {
		message.Stream = true
		return nil
	}
	/*
		That EOF error is because of the below code!
		Where, we are sending the data as the gob encoder way the
		Type mismatches and gob in the background converts to frames!
		And there is a chance the eof will come as read is done!
		The issue is whenever we are sending the data over the network we need to make sure
		the same type or the same way data has been read here to be decoded properly!
	*/
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	//If we remove the below line even the message from the server will not be decoded at the peer end.
	//Because the payload is not being properly treated as or being read in that payload of bytes way!
	message.Payload = buf[:n]

	return nil
}
