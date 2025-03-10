package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/sunnyashlesh16/go-distributedFileSystem/p2p"
	"io"
	"log"
	"sync"
)

type ServerOpts struct {
	RootStorageName string
	TransFunc       PathNameTransFunc
	Transport       p2p.Transport
	Network         []string
}

type Server struct {
	ServerOpts
	Store    *Store
	quitch   chan struct{}
	peerLock sync.Mutex
	peers    map[string]p2p.Peer
}

func NewFileServer(sopts ServerOpts) *Server {
	storeOpts := StoreOpts{
		Root:              sopts.RootStorageName,
		PathNameTransFunc: sopts.TransFunc,
	}
	return &Server{
		ServerOpts: sopts,
		Store:      NewStore(storeOpts),
		quitch:     make(chan struct{}),
		peers:      make(map[string]p2p.Peer),
	}
}

func (server *Server) Stop() {
	close(server.quitch)
}

/*  EVent Driven Programming Where the Select will do two things
    One is it will continuosly wait for the messages and listen to it and will not close the server!
	And the second one is if someone closes the channel that server is using then server.quitch will return a case which will
	return the function and closes the server and closes the transport too!
*/

type Payload struct {
	Key  string
	Data []byte
}

func (server *Server) loop() {
	defer func() {
		log.Println("Quitting the File Server Because Of the User Actions")
	_:
		server.Transport.Close()
	}()

	for {
		select {
		case msg := <-server.Transport.Queue():
			var p Payload
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&p); err != nil {
				log.Println("decoding error: ", err)
			}
			fmt.Print(p)
		case <-server.quitch:
			return
		}
	}
}

func (server *Server) broadcast(p *Payload) error {
	//Basically we are setting the empty peers as io writers
	/*
		The magic is when i try to save the data on to the store that
		is not sending the data to my peers when i click the multi writer to encode and send
		it to all my peers thats gonna reach my conn peer as an message
		which will trigger the handleConn functionalities and the message is set to the channel!
		As in the loop we are waiting to get the message from our transport channel!
		That will transferred!
		This is pure event driven and message queue structure
	*/
	peers := []io.Writer{}
	//Looping through each peers and append them to the peers
	for _, peer := range server.peers {
		peers = append(peers, peer)
	}
	//setting the multiwriter for ther peers
	mw := io.MultiWriter(peers...)
	//returning by encoding the mw wrt payload
	return gob.NewEncoder(mw).Encode(p)
}

func (server *Server) StoreData(key string, r io.Reader) error {
	//Get the Data & Store it

	buf := new(bytes.Buffer)
	//This will read the content from the reader and writes to buf and returns a reader
	tee := io.TeeReader(r, buf)
	//Using this reader we will be writing or storing the data
	if err := server.Store.Write(key, tee); err != nil {
		return err
	}
	//Now the payload is set properly!
	p := &Payload{Key: key, Data: buf.Bytes()}
	return server.broadcast(p)
}

func (server *Server) RegisterPeer(peer p2p.Peer) error {
	server.peerLock.Lock()
	defer server.peerLock.Unlock()
	server.peers[peer.RemoteAddr().String()] = peer
	fmt.Printf("Registering Peer %s For %s\n", peer.RemoteAddr(), server.RootStorageName)
	return nil
}

// This will Server the COnnections by dialing like a client
func (server *Server) Serve() error {
	for _, network := range server.Network {
		if len(network) == 0 {
			continue
		}
		go func(network string) {
			n := server.Transport.Call(network)
			if n != nil {
				fmt.Println("Connection Issue While Dialing")
			}
		}(network)
	}
	return nil
}

// This will get the requests from the listener
// As Of Now I belive TCP transport is a listener that waits for the connection or any new requests!and server deals with the process!
func (server *Server) Start() error {
	if err := server.Transport.ListenAndAccept(); err != nil {
		return err
	}

	server.Serve()
	server.loop()
	return nil
}
