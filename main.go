package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	inet "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

// create a 'Host' with a random peer to listen on the given address
func makeBasicHost(listen string, pid peer.ID) (host.Host, error) {
	addr, err := ma.NewMultiaddr(listen)
	if err != nil {
		return nil, err
	}

	ps := pstore.NewPeerstore()

	ctx := context.Background()

	// create a new swarm to be used by the service host
	netw, err := swarm.NewNetwork(ctx, []ma.Multiaddr{addr}, pid, ps, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("I am %s/ipfs/%s\n", addr, pid.Pretty())
	return bhost.New(netw), nil
}

func continueAsking() bool {
	fmt.Println("Do you have other peers to add? (true/false)")
	var shouldContinue bool

	n, err:= fmt.Scanln(&shouldContinue)
	if(n!=1 || err!=nil ){
		panic(err)
	}

	return shouldContinue
}

func main() {
	pids := make([]peer.ID, 100);
	streams := make([]inet.Stream, 100)

	fmt.Println("Hello, welcome to my simple p2p app.")

	fmt.Println("Please provide your prefered tcp port:")
	var myTcp string
	n, err:= fmt.Scanln(&myTcp)

	if(n!=1 || err!=nil ){
		panic(err)
	}

	fmt.Println("Provide your peerID")
	var myId string
	n, err= fmt.Scanln(&myId)

	if(n!=1 || err!=nil ){
		panic(err)
	}

	listenaddr := "/ip4/127.0.0.1/tcp/" +  myTcp

	myPId, err := peer.IDB58Decode(myId)
	if err != nil {
		log.Fatal(err)
	}
	pids = append(pids, myPId)

	ha, err := makeBasicHost(listenaddr, myPId)
	if err != nil {
		fmt.Println(listenaddr)
		log.Fatal(err)
	}

	// Set a stream handler on host A

	ha.SetStreamHandler("/p2pPublish/0.0.0", func(s net.Stream) {
		log.Println("Message on the stream")
		//defer s.Close()
		doWrite(s)
	})


	for ; continueAsking() ; {
		fmt.Println("Please provide your next peer's tcp port")
		var peerTcp string
		n, err:= fmt.Scanln(&peerTcp)

		if(n!=1 || err!=nil ){
			panic(err)
		}

		fmt.Println("Provide your next peer's id")
		var peerId string
		n, err= fmt.Scanln(&peerId)

		if(n!=1 || err!=nil ){
			panic(err)
		}

		peerid, err := peer.IDB58Decode(peerId)
		if err != nil {
			log.Fatal(err)
		}
		pids = append(pids, peerid)


		peerAddress := "/ip4/127.0.0.1/tcp/" + peerTcp + "/ipfs/" + peerId

		ipfsaddr, err := ma.NewMultiaddr(peerAddress)
		if err != nil {
			log.Fatalln(err)
		}

		tptaddr := strings.Split(ipfsaddr.String(), "/ipfs/")[0]
		// This creates a MA with the "/ip4/ipaddr/tcp/port" part of the target
		tptmaddr, err := ma.NewMultiaddr(tptaddr)
		if err != nil {
			log.Fatalln(err)
		}

		// We need to add the target to our peerstore, so we know how we can
		// contact it
		ha.Peerstore().AddAddr(peerid, tptmaddr, pstore.PermanentAddrTTL)

		fmt.Println("Press enter when ready")
		var ok string
		fmt.Scanln(&ok)

		s, err := ha.NewStream(context.Background(), peerid, "/p2pPublish/0.0.0")
		if err != nil {
			fmt.Println("chalut")
			log.Fatalln(err)
		}

		streams = append(streams, s)
	}

	for ; true ; {
		fmt.Println("Write down a message")

		var message string
		n, err:= fmt.Scanln(&message)

		if(n!=1 || err!=nil ){
			panic(err)
		}

		for _, s:= range streams {
			_, err = s.Write([]byte(message))
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}


func doWrite(s net.Stream) {

	buf := make([]byte, 1024)
	n, err := s.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("New message:")
	fmt.Println(buf[:n])


}

