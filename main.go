package main

import (
	"bytes"
	"github.com/sunnyashlesh16/go-distributedFileSystem/p2p"
	"log"
	"time"
)

func makeServer(listenAddr string, nodes ...string) *Server {
	tcpOpts := p2p.TCPTransportOps{
		ListenAddress: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	servOpts := ServerOpts{
		RootStorageName: listenAddr + "_files",
		TransFunc:       CASPathTransformFunc,
		Transport:       tr,
		Network:         nodes,
	}

	s := NewFileServer(servOpts)

	tr.OnPeer = s.RegisterPeer

	return s
}

func main() {

	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		if err := s1.Start(); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(1 * time.Second)
	go s2.Start()
	time.Sleep(1 * time.Second)
	data := bytes.NewReader([]byte("Hello World"))
	err := s2.StoreData("MyPrivateFolder", data)
	if err != nil {
		log.Fatal(err)
	}
}
