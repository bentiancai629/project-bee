package network

import (
	"fmt"
	"time"

	"project-bee/core"
	"project-bee/crypto"

	"github.com/sirupsen/logrus"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	RPCHandler RPCHandler
	Transports []Transport
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	blockTime   time.Duration
	memPool     *TxPool
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}

	s := &Server{
		ServerOpts:  opts,
		blockTime:   opts.BlockTime,
		memPool:     NewTxPool(),
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}
	if opts.RPCHandler == nil {
		opts.RPCHandler = NewDefaultRPCHandler(s)
	}

	s.ServerOpts = opts
	return s
}

func (s *Server) Start() {
	fmt.Println("Server start")

	s.initTransports()
	ticker := time.NewTicker(5 * time.Second)
free:
	for {
		select {
		case rpc := <-s.rpcCh:
			// buf := make([]byte, 10)
			// rpc.Payload.Read(buf)
			// fmt.Printf("From %s with Payload: %v \n", rpc.From, buf)
			if err := s.RPCHandler.HandleRPC(rpc); err != nil {
				logrus.Error(err)
			}
		case <-s.quitCh:
			break free
		case <-ticker.C:
			fmt.Println("do stuff every 5 seconds")
			if s.isValidator {
				s.createNewBlock()
			}
		}
	}

	fmt.Println("Server shutdown")
}

func (s *Server) handleTransaction(tx *core.Transaction) error {

	// 验证 tx 的 signature
	if err := tx.Verify(); err != nil {
		return err
	}

	hash := tx.Hash(core.TxHasher{})

	// 内存池
	logrus.WithFields(logrus.Fields{
		"hash": hash,
	}).Info("transaction already in mempool")
	if s.memPool.Has(hash) {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"hash": hash,
	}).Info("adding new tx to the mempool")

	return s.memPool.Add(tx)
}

func (s *Server) ProcessTransaction(from NetAddr, tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Has(hash) {
		logrus.WithFields(logrus.Fields{
			"hash": hash,
		}).Info("transaction already in mempool")

		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	logrus.WithFields(logrus.Fields{
		"hash":           hash,
		"mempool length": s.memPool.Len(),
	}).Info("adding new tx to the mempool")

	// TODO(@anthdm): broadcast this tx to peers

	return s.memPool.Add(tx)
}

func (s *Server) createNewBlock() error {
	fmt.Println("creating a new block")
	return nil
}

func (s *Server) initTransports() error {
	for _, tr := range s.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}
	return nil
}
