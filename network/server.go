package network

import (
	"bytes"
	"encoding/gob"
	"os"
	"time"

	"project-bee/core"
	"project-bee/crypto"
	"project-bee/types"

	"github.com/go-kit/log"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	TCPTransport  *TCPTransport
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	mempool     *TxPool
	chain       *core.Blockchain
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}

	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}

	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		// opts.Logger = log.With(opts.Logger, "addr", opts.Transport.Addr())
	}

	chain, err := core.NewBlockchain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}

	s := &Server{
		ServerOpts:  opts,
		chain:       chain,
		mempool:     NewTxPool(1000),
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}

	// If we dont got any processor from the server options, we going to use
	// the server as default.
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}

	if s.isValidator {
		go s.validatorLoop()
	}

	return s, nil

}
 
func (s *Server) Start() {
	// 并发启动监听
	s.TCPTransport.Start()
free:
	for {
		select {
		case rpc := <-s.rpcCh:

			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
			}

			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				// 过滤掉 已经同步的 err
				if err != core.ErrBlockKnown {
					s.Logger.Log("error", err)
				}
			}

		case <-s.quitCh:
			break free
			// case <-ticker.C:
			// 	if s.isValidator {
			// 		s.createNewBlock()
			// 	}
		}
	}

	s.Logger.Log("msg", "Server is shutting down")
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)

	s.Logger.Log("msg", "Starting validator loop", "blockTime", s.BlockTime)

	for {
		<-ticker.C
		s.createNewBlock()
	}
}

// 解析 Message 然后处理
func (s *Server) ProcessMessage(msg *DecodedMessage) error {
	// fmt.Printf("receiving message: %+v\n", msg.Data)
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		return s.processBlock(t)
	case *GetStatusMessage:
		// return s.processGetStatusMessage(msg.From, t)
	case *StatusMessage:
		// return s.processStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		// return s.processGetBlocksMessage(msg.From, t)
		// case *BlocksMessage:
		// 	return s.processBlocksMessage(msg.From, t)
	}

	return nil
}

// func (s *Server) processGetBlocksMessage(from NetAddr, data *GetBlocksMessage) error {
// 	s.Logger.Log("msg", "received getBlocks message", "from", from)

// 	var (
// 		blocks    = []*core.Block{}
// 		ourHeight = s.chain.Height()
// 	)

// 	if data.To == 0 {
// 		for i := int(data.From); i <= int(ourHeight); i++ {

// 		}
// 	}

// 	blocksMsg := &BlocksMessage{
// 		Blocks: blocks,
// 	}

// 	buf := new(bytes.Buffer)
// 	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
// 		return err
// 	}

// 	// s.mu.RLock()
// 	// defer s.mu.RUnlock()
// 	msg := NewMessage(MessageTypeBlocks, buf.Bytes())
// 	return s.Transport.SendMessage(from, msg.Bytes())
// }

func (s *Server) processBlocksMessage(from NetAddr, data *BlocksMessage) error {
	return nil
}

// func (s *Server) processStatusMessage(from NetAddr, data *StatusMessage) error {
// 	if data.CurrentHeight > s.chain.Height() {
// 		s.Logger.Log("msg", "cannot sync blockHeight to low", "ourHeight", s.chain.Height(), "theirHeight", data.CurrentHeight, "addr", from)
// 		return nil
// 	}

// 	getBlocksMessage := &GetBlocksMessage{
// 		From: s.chain.Height(),
// 		To:   0,
// 	}

// 	buf := new(bytes.Buffer)
// 	if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
// 		return err
// 	}

// 	msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
// 	return s.Transport.SendMessage(from, msg.Bytes())
// }

// TODO
// func (s *Server) processGetStatusMessage(from NetAddr, data *GetStatusMessage) error {
// 	s.Logger.Log("msg", "received getStatus message", "from", from)

// 	statusMessage := &StatusMessage{
// 		CurrentHeight: s.chain.Height(),
// 		ID:            s.ID,
// 	}

// 	buf := new(bytes.Buffer)
// 	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
// 		return err
// 	}

// 	// s.mu.RLock()
// 	// defer s.mu.RUnlock()

// 	// peer, ok := s.peerMap[from]
// 	// if !ok {
// 	// 	return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
// 	// }

// 	msg := NewMessage(MessageTypeStatus, buf.Bytes())

// 	return s.Transport.SendMessage(from, msg.Bytes())
// 	// return peer.Send(msg.Bytes())
// 	// return nil
// }

func (s *Server) processBlock(b *core.Block) error {
	if err := s.chain.AddBlock(b); err != nil {
		return err
	}

	go s.broadcastBlock(b)

	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	if s.mempool.Contains(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	// s.Logger.Log(
	// 	"msg", "adding new tx to mempool",
	// 	"hash", hash.ToHexString(),
	// 	"mempoolLength", s.memPool.Len(),
	// )

	go s.broadcastTx(tx)
	s.mempool.Add(tx)

	return nil
}

// 节点之间同步高度
func (s *Server) sendGetStatusMessage(tr Transport) error {
	var (
		/*
			ID            string
			Version       uint32
			CurrentHeight uint32
		*/
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)
	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}

	// 节点之间同步 message
	// msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	// if err := s.Transport.SendMessage(tr.Addr(), msg.Bytes()); err != nil {
	// 	return err
	// }

	return nil
}

// 广播 message	 到所有节点
func (s *Server) broadcast(payload []byte) error {
	// for _, tr := range s.Transports {
	// 	if err := tr.Broadcast(payload); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeBock, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeTx, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	// For now we are going to use all transactions that are in the mempool
	// Later on when we know the internal structure of our transaction
	// we will implement some kind of complexity function to determine how
	// many transactions can be included in a block.
	txs := s.mempool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txs)
	if err != nil {
		return err
	}

	if err := block.Sign(*s.PrivateKey); err != nil {
		return err
	}

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	s.mempool.ClearPending()

	go s.broadcastBlock(block)

	return nil
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Height:    0,
		Timestamp: 000000,
	}

	b, _ := core.NewBlock(header, nil)
	return b
}
