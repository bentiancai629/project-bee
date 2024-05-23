package main

import (
	"bytes"
	"encoding/gob"

	"log"
	"net/http"
	"time"

	"project-bee/core"
	"project-bee/crypto"
	"project-bee/network"
)

func main() {
	privKey := crypto.GeneratePrivateKey()

	localNode := makeServer("LOCAL_NODE", &privKey, ":3000", []string{":4000"}, ":9000")
	go localNode.Start()

	remoteNode := makeServer("REMOTE_NODE", nil, ":4000", []string{":5000"}, "")
	go remoteNode.Start()

	remoteNodeB := makeServer("REMOTE_NODE_B", nil, ":5000", nil, "")
	go remoteNodeB.Start()

	go func() {
		time.Sleep(11 * time.Second)

		lateNode := makeServer("LATE_NODE", nil, ":6000", []string{":4000"}, "")
		go lateNode.Start()
	}()

	time.Sleep(1 * time.Second)

	txSenderTicker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-txSenderTicker.C:
				txSender()
			}
		}
	}()

	select {}

}

func makeServer(id string, pk *crypto.PrivateKey, addr string, seedNodes []string, apiListenAddr string) *network.Server {
	opts := network.ServerOpts{
		APIListener: apiListenAddr,
		SeedNodes:   seedNodes,
		ListenAddr:  addr,
		PrivateKey:  pk,
		ID:          id,
	}

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func txSender() {
	// conn, err := net.Dial("tcp", ":3000")
	// if err != nil {
	// 	panic(err)
	// }

	privKey := crypto.GeneratePrivateKey()

	// data := []byte("Hello BTC!")
	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}
	// msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	// _, err = conn.Write(msg.Bytes())
	// if err != nil {
	// 	panic(err)
	// }

	req, err := http.NewRequest("POST", "http://localhost:9000/tx", buf)
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
}

// var transports = []network.Transport{
// 	network.NewLocalTransport("LOCAL"),
// 	// network.NewLocalTransport("REMOTE_A"),
// 	// network.NewLocalTransport("REMOTE_B"),
// 	// network.NewLocalTransport("REMOTE_C"),
// }

// func main1() {

// 	// node A/B/C
// 	initRemoteServers(transports)

// 	localNode := transports[0]
// 	trLate := network.NewLocalTransport("LATE_REMOTE")
// 	// remoteNodeA := transports[1]
// 	// remoteNodeC := transports[3]

// 	// go func() {
// 	// 	for {
// 	// 		if err := sendTransaction(remoteNodeA, localNode.Addr()); err != nil {
// 	// 			logrus.Error(err)
// 	// 		}
// 	// 		time.Sleep(2 * time.Second)
// 	// 	}
// 	// }()

// 	// node Late
// 	go func() {
// 		time.Sleep(7 * time.Second)
// 		lateServer := makeServer(string(trLate.Addr()), trLate, nil)
// 		go lateServer.Start()
// 	}()

// 	privKey := crypto.GeneratePrivateKey()
// 	localServer := makeServer("LOCAL", localNode, &privKey)
// 	localServer.Start()
// }

// func initRemoteServers(trs []network.Transport) {
// 	for i := 0; i < len(trs); i++ {
// 		id := fmt.Sprintf("REMOTE_%d", i)
// 		s := makeServer(id, trs[i], nil)
// 		go s.Start()
// 	}
// }

func sendGetStatusMessage(tr network.Transport, to network.NetAddr) error {
	var (
		/*
			ID            string
			Version       uint32
			CurrentHeight uint32
		*/
		getStatusMsg = new(network.GetStatusMessage)
		buf          = new(bytes.Buffer)
	)
	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}

	// 节点之间同步 message
	msg := network.NewMessage(network.MessageTypeGetStatus, buf.Bytes())
	return tr.SendMessage(tr.Addr(), msg.Bytes())

}

// func sendTransaction(tr network.Transport, to network.NetAddr) error {
// 	privKey := crypto.GeneratePrivateKey()
// 	// data := []byte(strconv.FormatInt(int64(rand.Intn(1000000000)), 10))
// 	// data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}
// 	// contract := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}
// 	data := []byte{0x03, 0x0a}
// 	tx := core.NewTransaction(data)
// 	tx.Sign(privKey)
// 	buf := &bytes.Buffer{}
// 	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
// 		return err
// 	}

// 	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

// 	return tr.SendMessage(to, msg.Bytes())
// }
