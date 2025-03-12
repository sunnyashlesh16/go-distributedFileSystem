package main

import (
	"fmt"
	"github.com/sunnyashlesh16/go-distributedFileSystem/p2p"
	"io"
	"log"
	"strings"
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
		RootStorageName: strings.Split(listenAddr, ":")[1] + "_network",
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

	//data := bytes.NewReader([]byte("This is the largest File"))
	//s2.StoreData("MySecond Folder", data)

	r, err := s2.GetData("MySecond Folder")
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
	select {}
}
