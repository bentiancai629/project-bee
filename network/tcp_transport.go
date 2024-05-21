package network

import (
	"fmt"
	"io"
	"net"
)

type TCPPeer struct {
	conn     net.Conn
	Outgoing bool
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err
}

type TCPTransport struct {
	peerCh     chan *TCPPeer
	listenAddr string
	listener   net.Listener
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
	}
}

func (t *TCPTransport) readLoop(peer *TCPPeer) {
	buf := make([]byte, 2048)
	for {
		n, err := peer.conn.Read(buf)
		if err == io.EOF {
			continue
		}
		if err != nil {
			fmt.Printf("read error: %s", err)
			continue
		}

		msg := buf[:n]
		fmt.Println("read msg: ", string(msg))
		// rpcCh <- RPC{
		// 	From:    p.conn.RemoteAddr(),
		// 	Payload: bytes.NewReader(msg),
		// }
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

		// fmt.Printf("new incoming TCP connection =>%+v\n", conn)

		go t.readLoop(peer)
		// t.peerCh <- peer
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