package p2p

import (
	"fmt"
	"io"
	"net"
)

type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{conn: conn, outbound: outbound}
}

func (peer *TCPPeer) Close() error {
	return peer.conn.Close()
}

type TCPTransportOps struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOps
	listener     net.Listener
	queuemessage chan RPC
}

func NewTCPTransport(opts TCPTransportOps) *TCPTransport {
	return &TCPTransport{
		TCPTransportOps: opts,
		queuemessage:    make(chan RPC),
	}
}

func (transport *TCPTransport) Queue() <-chan RPC {
	return transport.queuemessage
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %s\n", err)
		}
		fmt.Printf("Accepting New connection from %s\n", conn.RemoteAddr())
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	//This defer function make sure when the execution of the current function ends it will call this function!
	defer func() {
		fmt.Printf("Closing connection: %s\n", conn.RemoteAddr())
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)
	//If a value is not being assigned with (:) this then its using the already defined var!
	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		//This OnPeer is a callback function that we are using as a notification when a peer is connected!
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	msg := RPC{}
	for {
		err := t.Decoder.Decode(conn, &msg)
		if err == net.ErrClosed || err == io.EOF {
			return
		}

		/*
			There was A error here as we know when a conncetion is closed from the client side or in some way
			we were handling the error based connection closing like No Peer Connection or no handshake successful
			But with the client closing the connection i think that can be handled using io.EOF or ErrClosed
			Which is fixed now!
			We can use panic to print what kind of error its!
		*/
		if err != nil {
			fmt.Printf("Error Reading the message or decoding to peer: %s\n", err)
			continue
		}

		msg.From = conn.RemoteAddr()
		t.queuemessage <- msg
	}
}
