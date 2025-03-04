package p2p

import (
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP established connection.
//type TCPPeer struct {
//	// The underlying connection of the peer. Which in this case
//	// is a TCP connection.
//	conn net.Conn
//	// if we dial and retrieve a conn => outbound == true
//	// if we accept and retrieve a conn => outbound == false
//	outbound bool
//}
//
//func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
//	return &TCPPeer{
//		Conn:     conn,
//		outbound: outbound,
//	}
//}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
	}
}
