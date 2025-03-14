package main

import (
	"bytes"
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
		Key:             newEncryptionKey(),
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
	s2 := makeServer(":4000", "")
	s3 := makeServer(":5000", ":4000", ":3000")

	go func() {
		if err := s1.Start(); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(5 * time.Millisecond)
	go func() {
		if err := s2.Start(); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(2 * time.Second)
	go s3.Start()
	time.Sleep(2 * time.Second)
	//for i := 1; i < 11; i++ {
	//	data := bytes.NewReader([]byte(fmt.Sprintf("This is the the data of my fileNumber-%d", i)))
	//_:
	//	s2.StoreData(fmt.Sprintf("MyFileNumber-%d", i), data)
	//	time.Sleep(5 * time.Millisecond)
	//}
	for i := 1; i < 2; i++ {
		key := fmt.Sprintf("EncryptedFile-%d", i)
		data := bytes.NewReader([]byte(fmt.Sprintf("This is the largest File-%d", i)))
		err := s3.StoreData(key, data)
		if err != nil {
			log.Fatal(err)
		}

		if err := s3.Store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}
		time.Sleep(2 * time.Second)
		r, err := s3.GetData(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	}
}

// TODO: Handle Delete Properly In This Flow
// TODO: Add Delete & Update APIs
// TODO: Clean UP & Refactor Code & Add Debugging Statements & Write Clear Function Definition & Add a Diagram
