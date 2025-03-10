package p2p

import (
	"errors"
	"fmt"
	"io"
	"net"
)

type TCPPeer struct {
	net.Conn
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{Conn: conn, outbound: outbound}
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

func (tcppeer *TCPPeer) Send(b []byte) error {
	_, err := tcppeer.Conn.Write(b)
	return err
}

// Closes the TCP Connection Socket
func (transport *TCPTransport) Close() error {

	return transport.listener.Close()
}

func (transport *TCPTransport) Queue() <-chan RPC {
	return transport.queuemessage
}

func (transport *TCPTransport) Call(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go transport.handleConn(conn, true)
	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	fmt.Println("Listening on " + t.ListenAddress)
	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("Error accepting connection: %s\n", err)
		}
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	//This defer function make sure when the execution of the current function ends it will call this function!
	defer func() {
		fmt.Printf("Closing connection: %s\n", conn.RemoteAddr())
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)
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
