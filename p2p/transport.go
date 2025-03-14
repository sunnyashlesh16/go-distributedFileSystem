package p2p

import "net"

type Transport interface {
	Addr() string
	ListenAndAccept() error
	Queue() <-chan RPC
	Close() error
	Call(addr string) error
}

type Peer interface {
	net.Conn
	Send(b []byte) error
	Close() error
	Closestream()
}
