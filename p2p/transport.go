package p2p

type Transport interface {
	ListenAndAccept() error
	Queue() <-chan RPC
}

type Peer interface {
	Close() error
}
