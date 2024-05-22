package network

import (
	"fmt"
	"io"
	"net"
	"bytes"
)

type TCPPeer struct {
	conn     net.Conn
	Outgoing bool
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err 
}

// TODO 
// read error: read tcp 127.0.0.1:59481->127.0.0.1:5000: read: connection reset by peeraddr=LOCAL_NODE msg="new block" hash=4907e5bb2dec7e62cce104ec8f844939735a8bb72dda57ff7054604f9166c3ce height=9 transactions=0
func (p *TCPPeer) readLoop(rpcCh chan RPC) {
	buf := make([]byte, 4096)
	for {
		n, err := p.conn.Read(buf)
		if err == io.EOF {
			continue
		}
		if err != nil {
			fmt.Printf("read error: %s", err)
			continue
		}

		msg := buf[:n]

		// fmt.Println("read msg: ", string(msg))

		rpcCh <- RPC{
			From:    p.conn.RemoteAddr(),
			Payload: bytes.NewReader(msg),
		}
	}
}

type TCPTransport struct {
	peerCh     chan *TCPPeer
	listenAddr string
	listener   net.Listener
}

func NewTCPTransport(addr string, peerCh chan *TCPPeer) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
		peerCh:     peerCh,
	}
}

func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		peer := &TCPPeer{
			conn: conn,
		}

		t.peerCh <- peer

		// fmt.Printf("new incoming TCP connection =>%+v\n", conn)

		// go t.readLoop(peer)

	}
}

func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = ln

	go t.acceptLoop() //并发多个监听

	fmt.Println("TCP Transport listening to port: ", t.listenAddr)
	return nil
}
