/*
Copyright (c) 2018 TrueChain Foundation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pbftserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"path"
	"time"

	"pbft-core"
	pb "pbft-core/fastchain"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc"
)

// PbftServer defines the base properties of a pbft node server
type PbftServer struct {
	IP      string
	Port    int
	Nd      *pbft.Node
	Cfg     *pbft.Config
	Out     chan pbft.ApplyMsg
	TxnPool chan pb.Transaction
	Tc      *pb.TrueChain
}

type fastChainServer struct {
	pbftSv *PbftServer
}

// Start - Initial server logic goes here
func (sv *PbftServer) Start() {
	pbft.MyPrint(1, "Firing up peer server...\n")
}

func (sv *PbftServer) verifyTxnReq(req *pb.Transaction) bool {
	sig := req.Data.Signature
	pubkey, _ := ethcrypto.Ecrecover(req.Data.Hash, sig)

	clientPubKeyFile := fmt.Sprintf("sign%v.pub", sv.Cfg.N)
	fmt.Println("fetching file: ", clientPubKeyFile)
	clientPubKey, _ := pbft.FetchPublicKeyBytes(path.Join(sv.Cfg.KD, clientPubKeyFile))

	if bytes.Equal(pubkey, clientPubKey) {
		return true
	}

	return false
}

// createInternalPbftReq wraps a transaction request from client for internal rpc communication between pbft nodes
func (sv *PbftServer) createInternalPbftReq(proposedBlock *pb.PbftBlock) pbft.Request {
	req := pbft.Request{}
	reqInner := pbft.RequestInner{}

	reqInner.ID = sv.Cfg.N // client-id
	reqInner.Seq = 0
	reqInner.View = 0
	reqInner.Reqtype = pbft.TypeRequest // client request
	reqInner.Block = proposedBlock
	reqInner.Timestamp = time.Now().Unix()

	req.Inner = reqInner

	req.AddSig(sv.Nd.EcdsaKey)

	return req
}

// NewTrueChain creates a fresh blockchain
func (sv *PbftServer) NewTrueChain() {
	genesisBlock := pbft.GetDefaultGenesisBlock()
	pbft.MyPrint(0, "Genesis block generated: %x\n\n", genesisBlock.Header.TxnsHash)

	tc := &pb.TrueChain{}

	tc.Blocks = make([]*pb.PbftBlock, 0)
	tc.Blocks = append(tc.Blocks, genesisBlock)
	tc.LastBlockHeader = genesisBlock.Header

	sv.Tc = tc
}

func (sv *PbftServer) AppendBlock(block *pb.PbftBlock) {
	sv.Tc.Blocks = append(sv.Tc.Blocks, block)
	sv.Tc.LastBlockHeader = block.Header
}

func (sv *PbftServer) addToTxnPool(txn pb.Transaction) {
	sv.TxnPool <- txn
}

// NewTxnRequest handles transaction rquests from clients
func (sv *fastChainServer) NewTxnRequest(ctx context.Context, txnReq *pb.Transaction) (*pb.GenericResp, error) {
	if sv.pbftSv.verifyTxnReq(txnReq) {
		fmt.Println("Txn verified")
	} else {
		fmt.Println("Txn verification failed")
		return &pb.GenericResp{Msg: "Transaction verification failed"}, errors.New("Invalid transaction request")
	}

	sv.pbftSv.addToTxnPool(*txnReq)
	return &pb.GenericResp{Msg: fmt.Sprintf("Transaction request received in node %d\n", sv.pbftSv.Nd.ID)}, nil
}

// RegisterPbftGrpcListener listens to client for new transaction requests on grpcPort
func RegisterPbftGrpcListener(grpcPort int, sv *PbftServer) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterFastChainServer(grpcServer, &fastChainServer{pbftSv: sv})
	go grpcServer.Serve(lis)
}

// BuildServer initiates the Server resource properties and listens to client's
// message requests as well as interacts with the channel
func BuildServer(cfg pbft.Config, IP string, port int, grpcPort int, me int) *PbftServer {
	sv := &PbftServer{}
	sv.IP = IP
	sv.Port = port
	sv.Out = make(chan pbft.ApplyMsg, cfg.NumQuest)
	sv.Cfg = &cfg
	sv.TxnPool = make(chan pb.Transaction)

	applyChan := make(chan pbft.ApplyMsg, cfg.NumQuest)
	sv.Nd = pbft.Make(cfg, me, port, 0, applyChan, 100) // test 100 messages

	RegisterPbftGrpcListener(grpcPort, sv)

	sv.NewTrueChain()

	//if sv.Nd.ID == sv.Nd.Primary {
	// If this is a primary replica/leader extract txns from txnpool
	// in batches, create blocks and broadcast to other nodes
	blockTxnChan := make(chan pb.Transaction, cfg.Blocksize)

	// This goroutine keeps extracting transactions from txnpool
	// and adds it to buffered channel of length block-size
	go func(blockTxnChan chan pb.Transaction) {
		for {
			txn := <-sv.TxnPool
			blockTxnChan <- txn
		}
	}(blockTxnChan)

	// This ensures that new block isn't sent through pbft phases unless the current block has been committed
	// TODO use wg.WaitGroup() in the future
	committed := make(chan bool, 1)
	committed <- true

	// This go routine creates the slices of block-size length and creates the block
	// for broadcasting once the previous block has been committed
	go func() {
		for {
			blockTxns := make([]*pb.Transaction, 0)
			var gasUsed int64
			gasUsed = 0
			for i := 0; i <= cfg.Blocksize; i++ {
				txn := <-blockTxnChan
				blockTxns = append(blockTxns, &txn)
				pbft.MyPrint(1, "Adding transaction request %s to block", string(txn.Data.Payload))
				gasUsed = gasUsed + txn.Data.Price
			}
			parentHash := pbft.HashBlockHeader(sv.Tc.LastBlockHeader)
			txnsHash := pbft.HashTxns(blockTxns)
			header := pbft.NewPbftBlockHeader(sv.Tc.LastBlockHeader.Number+1, 5000, int64(gasUsed), parentHash, txnsHash)

			block := pbft.NewPbftBlock(header, blockTxns)

			<-committed
			req := sv.createInternalPbftReq(block)
			sv.Nd.NewClientRequest(req, cfg.N)
		}
	}()
	//}

	go func(aC chan pbft.ApplyMsg) {
		for {
			c := <-aC
			pbft.MyPrint(4, "[0.0.0.0:%d] [%d] New Sequence Item: %v\n", sv.Port, me, c)
			sv.Out <- c
			block := <-sv.Nd.CommittedBlock
			sv.AppendBlock(block)
			committed <- true
		}
	}(applyChan)

	go sv.Start() // in case the server has some initial logic
	return sv
}
