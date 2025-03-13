package p2p

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

type TCPPeer struct {
	net.Conn
	outbound bool
	//If can use capital one as the starting letter then it can be exported !
	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{Conn: conn, outbound: outbound, wg: &sync.WaitGroup{}}
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

func (transport *TCPTransport) Addr() string {
	return transport.ListenAddress
}

func (peer *TCPPeer) Closestream() {
	peer.wg.Done()
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

	for {
		fmt.Println("Accepting the Messages Continuously Over the Connection")
		msg := RPC{}
		fmt.Println("Setting The Message")
		//This will Accept the data using the r from the conn and align with the msg!
		err := t.Decoder.Decode(conn, &msg)
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
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
		/*
			Here is the crazy part where we know the normal messages or small messages will be directly sent to the
			connection and can be read using event driven architecture we have set up in the loop using channel triggers
			encoding, and decoding way with proper type

			But When it comes to large data or streaming or uploading or reading we need the help of chunking,
			go routines to divide, and wait groups/locks!
			When ever file is uploading we can set a go routine for each chunk with a wait group and make sure we defer once that chunk
			is done!
			Before saying the uploading is done! we will set wait group as wait!
			And once all the chunks are done we will set wait group as done  to print the task is done
			So when there are mutiple wg's,
			wg.add will add a thread lets say only when in a thread wg.Done is called the wg will be decreased!
			but wg.wait will wait until the wg is set to 0!

			But how can the difference be made between small and large!
			A control frame with 16 byte frame which will have the frame type so !
			That frame type will be 0xO1 or 0x02?
		*/
		fmt.Printf("Received message but this is at transport end from %s: of %s\n", conn.RemoteAddr(), msg)
		msg.From = conn.RemoteAddr().String()
		if msg.Stream {
			fmt.Println("Peer is added to wait group")
			peer.wg.Add(1)
			fmt.Println("Peer is waiting")
			peer.wg.Wait()
			fmt.Println("Peer is Removed From Waiting Group Upon Calling the done")
			continue
		}
		fmt.Println("Message Is Added to the channel")
		t.queuemessage <- msg
		//fmt.Printf("Peer: %s is removed from wg and for loop can normally accept the messages\n", peer.RemoteAddr().String())
	}
}
