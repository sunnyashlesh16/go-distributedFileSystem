package p2p

import "net"

type Transport interface {
	ListenAndAccept() error
	Queue() <-chan RPC
	Close() error
	Call(addr string) error
}

type Peer interface {
	net.Conn
	Close() error
}
