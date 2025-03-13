package main

import (
	"bytes"
	"fmt"
	"github.com/sunnyashlesh16/go-distributedFileSystem/p2p"
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

	for i := 1; i < 11; i++ {
		data := bytes.NewReader([]byte(fmt.Sprintf("This is the the data of my fileNumber-%d", i)))
	_:
		s2.StoreData(fmt.Sprintf("MyFileNumber-%d", i), data)
		time.Sleep(5 * time.Millisecond)
	}

	//data := bytes.NewReader([]byte("This is the largest File"))
	//s2.StoreData("PersonalFile", data)

	//r, err := s2.GetData("PersonalFile")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//b, err := io.ReadAll(r)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println(string(b))
}
