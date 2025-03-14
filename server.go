package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/sunnyashlesh16/go-distributedFileSystem/p2p"
	"io"
	"log"
	"sync"
	"time"
)

// For anything we are setting up for any, In the case of message we are setting the apyload as any!
// So, we need to register!
func init() {
	gob.Register(MessageStore{})
	gob.Register(MessageGetFile{})
}

type ServerOpts struct {
	ID              string
	Key             []byte
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
	if len(sopts.ID) == 0 {
		sopts.ID = GenerateID()
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

/*
First I will Send the message With The Key
At Transport it will be decoded the encoded key name and then it will send the message to a rpc channel1
At the loop it will receive through the rpc and calls the handle message!
Up until at the tcp connection will be on hold to not receive any more messages!
Now, When another data which is basically streaming the data to the peer is done thorugh io copy!
As this is a direct raw data stream, when we call the write there will be no need of handling the data here!
Use the peer read connection it will fetch all the data and map that with the key!
Now the key will be hashed to store on the server!
With the data beimng copied from the raw peer connection to the file! And Saved!
Now the wg is called off and ready to server the next messages!
There is a bug here where the io.Copy in the write is being waiting to read the data!
So, we will limit it !Based on the size i mean the reader !
So that loop will be done!
*/
type Message struct {
	Payload any
}

type MessageStore struct {
	ID   string
	Key  string
	Size int64
}

type MessageGetFile struct {
	ID  string
	Key string
}

/*
	    EVent Driven Programming Where the Select will do two things
	    One is it will continuosly wait for the messages and listen to it and will not close the server!
		And the second one is if someone closes the channel that server is using then server.quitch will return a case which will
		return the function and closes the server and closes the transport too!
*/
func (server *Server) loop() {
	defer func() {
		log.Println("Quitting the File Server Because Of the User Actions")
	_:
		server.Transport.Close()
	}()

	for {
		select {
		case rpc := <-server.Transport.Queue():
			message := Message{}
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&message); err != nil {
				log.Printf("decoding error:%v", err)
			}
			fmt.Printf("Printing payload:%s at the peer loop end\n", message)
			if err := server.handleMessage(rpc.From, &message); err != nil {
				log.Printf("handle message error:%s", err)
			}
		case <-server.quitch:
			return
		}
	}
}

func (server *Server) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStore:
		fmt.Printf("Received Store Message For the File:%s\n", v.Key)
		if err := server.handleMessageStore(from, v); err != nil {
			log.Println("handle message error: ", err)
		}

	case MessageGetFile:
		fmt.Printf("Received Get Message For the File:%s\n", v.Key)
		if err := server.handleMessageGetFile(from, v); err != nil {
			log.Println("handle message error: ", err)
		}
	}
	return nil
}

func (server *Server) handleMessageGetFile(from string, msg MessageGetFile) error {
	fmt.Println("Need TO Get A File From Disk and Send it oVer the wire This is peer")
	fmt.Printf("Error ikkaada:%s\n", msg.ID)
	if !server.Store.HasFile(msg.ID, msg.Key) {
		return errors.New("file Doesn't Exists in network")
	}

	fmt.Println("Reading The File As If its in the Store")
	r, fileSize, err := server.Store.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}
	if rc, ok := r.(io.ReadCloser); ok {
		defer func() {
			_ = rc.Close()
		}()
	}

	time.Sleep(5 * time.Millisecond)

	peer, ok := server.peers[from]
	if !ok {
		return errors.New("Peer Doesn't Exists")
	}
	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, &fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("Written The File The peers %s", n)

	return nil
}

func (server *Server) handleMessageStore(from string, msg MessageStore) error {
	peer, ok := server.peers[from]
	if !ok {
		log.Println("peer not found")
	}
	fmt.Printf("Received Data Message:%v At Storing To The Peer:%s\n", msg, peer)
	// Limiting the reader to wait only for 25 bytes during the streaming!
	if _, err := server.Store.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size)); err != nil {
		log.Println("writing error: ", err)
	}
	fmt.Printf("Stored The File From The Peer:%s Locally\n", peer)
	fmt.Println("Calling Off the Waiting Group")
	peer.Closestream()
	return nil
}

func (server *Server) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	fmt.Printf("Printing the Message: %s before broadcasting to the peers\n", msg)
	for _, peer := range server.peers {
		fmt.Printf("Sending Message Byte To Peer: %s\n", peer)
	_:
		peer.Send([]byte{p2p.IncomingMessage})
		fmt.Printf("Sending The Message:%s To the peer:%s\n", msg, peer)
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (server *Server) GetData(key string) (io.Reader, error) {
	if b := server.Store.HasFile(server.ID, key); b != false {
		fmt.Println("Getting Data- Found in the Local")
		r, _, _ := server.Store.Read(server.ID, key)
		return r, nil
	}

	fmt.Println("Don't have the File Locally So We are Gonna Get it from network")

	msg := Message{MessageGetFile{ID: server.ID, Key: hashKey(key)}}
	fmt.Printf("Printing the Message ID Before Sending: %s\n", msg.Payload.(MessageGetFile).ID)

	if err := server.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(5 * time.Millisecond)

	for _, peer := range server.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := server.Store.WriteEncrypt(server.Key, server.ID, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("Writing Data The peers %s", n)
		fmt.Printf("Got The Data From Peers: %s\n", msg)
		peer.Closestream()
	}
	//select {}
	re, _, _ := server.Store.Read(server.ID, key)
	return re, nil
}

func (server *Server) StoreData(key string, r io.Reader) error {

	var (
		//Get the Data & Store it
		ownBuf = new(bytes.Buffer)
		//Right Once the r is read there won be data!Sp, using Tee reader
		tee = io.TeeReader(r, ownBuf)
	)

	n, err := server.Store.Write(server.ID, key, tee)
	if err != nil {
		return err
	}
	fmt.Printf("Stored the data of file: %s on to the own server before broadcasting to the peers\n", key)

	msg := Message{
		Payload: MessageStore{ID: server.ID, Key: hashKey(key), Size: n + 16},
	}

	if err := server.broadcast(&msg); err != nil {
		return err
	}
	fmt.Println("BroadCasted The File Message to The Peers")

	time.Sleep(5 * time.Millisecond)

	peers := []io.Writer{}
	for _, peer := range server.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)

	mw.Write([]byte{p2p.IncomingStream})
	//fmt.Printf(" Streaming The Data:%s to the peer %s\n", msg, peer)
	_, err = CopyEncrypt(server.Key, ownBuf, mw)
	if err != nil {
		return err
	}

	fmt.Printf("%d are Written to peers TCP Raw Socket\n", n)
	return nil
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
