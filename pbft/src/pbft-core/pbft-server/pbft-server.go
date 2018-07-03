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
	"errors"
	"fmt"
	"log"
	"net"
	"path"
	"time"

	"pbft-core"
	pb "pbft-core/fastchain"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/net/context"
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
	return &pb.GenericResp{Msg: "Transaction request received"}, nil
}

// BuildServer initiates the Server resource properties and listens to client's
// message requests as well as interacts with the channel
func BuildServer(cfg pbft.Config, IP string, Port int, GrpcPort int, me int) *PbftServer {
	sv := &PbftServer{}
	sv.IP = IP
	sv.Port = Port
	sv.Out = make(chan pbft.ApplyMsg, cfg.NumQuest)
	sv.Cfg = &cfg
	sv.TxnPool = make(chan pb.Transaction)

	blockTxnChan := make(chan pb.Transaction, cfg.Blocksize)
	go func(blockTxnChan chan pb.Transaction) {
		for {
			txn := <-sv.TxnPool
			blockTxnChan <- txn
		}
	}(blockTxnChan)

	go func() {
		for {
			blockTxns := make([]*pb.Transaction, 0)
			for i := 0; i <= cfg.Blocksize; i++ {
				txn := <-blockTxnChan
				blockTxns = append(blockTxns, &txn)
				fmt.Println("Adding transaction request %s to block", string(txn.Data.Payload))
			}

			pbftBlock := pb.PbftBlock{}
			pbftBlock.Header = &pb.PbftBlockHeader{}
			pbftBlock.Txns = blockTxns
			fmt.Println("Block created")

			req := sv.createInternalPbftReq(&pbftBlock)
			sv.Nd.NewClientRequest(req, cfg.N)
		}
	}()

	applyChan := make(chan pbft.ApplyMsg, cfg.NumQuest)

	go func(aC chan pbft.ApplyMsg) {
		for {
			c := <-aC
			//pbft.MyPrint(4, "[0.0.0.0:%d] [%d] New Sequence Item: %v\n", sv.Port, me, c)
			sv.Out <- c
		}
	}(applyChan)
	sv.Nd = pbft.Make(cfg, me, Port, 0, applyChan, 100) // test 100 messages

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", GrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterFastChainServer(grpcServer, &fastChainServer{pbftSv: sv})
	go grpcServer.Serve(lis)

	go sv.Start() // in case the server has some initial logic
	return sv
}
