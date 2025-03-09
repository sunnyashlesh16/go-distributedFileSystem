package main

import (
	"fmt"
	"github.com/sunnyashlesh16/go-distributedFileSystem/p2p"
	"log"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

// Function Reference
func OnPeer(peer p2p.Peer) error {
	fmt.Println("Peer is Connected")
	return nil
}

func main() {
	tcpOpts := p2p.TCPTransportOps{
		ListenAddress: ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	//Go Routine For Message Processing
	go func() {
		for {
			msg := <-tr.Queue()
			fmt.Printf("%+v\n", msg)
		}
	}()

	select {}
}
