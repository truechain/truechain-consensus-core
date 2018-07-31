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

package pbft

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"trueconsensus/common"

	pb "trueconsensus/fastchain/proto"

	"google.golang.org/grpc"
	// "trueconsensus/fastchain"
)

// Server defines the base properties of a pbft node server
type Server struct {
	IP        string
	Port      int
	Nd        *Node
	Cfg       *Config
	committed chan bool
}

type fastChainServer struct {
	pbftSv *Server
}

// Start - Initial server logic goes here
func (sv *Server) Start() {
	common.MyPrint(1, "Firing up peer server...\n")
}

// NewTxnRequest handles transaction requests from clients
func (sv *fastChainServer) NewTxnRequest(ctx context.Context, txnReq *pb.Transaction) (*pb.GenericResp, error) {
	common.MyPrint(4, "Received new transacion request %d from client on node %d", txnReq.Data.AccountNonce, sv.pbftSv.Nd.ID)

	// Discard already known transactions
	if sv.pbftSv.Nd.txPool.all.Get(common.BytesToHash(txnReq.Data.Hash)) != nil {
		return &pb.GenericResp{Msg: "Already known transaction. Ignoring transaction request."}, errors.New("Transaction already exists in pool")
	}

	sender, ok := VerifySender(txnReq, sv.pbftSv.Cfg.Network.N)
	if ok {
		common.MyPrint(0, "Transaction sender verified")
	} else {
		common.MyPrint(0, "Transaction verification failed")
		return &pb.GenericResp{Msg: "Transaction verification failed"}, errors.New("Invalid transaction request")
	}

	sv.pbftSv.Nd.txPool.Add(txnReq, sender)
	common.MyPrint(4, "Added request %d to transaction pool on node %d.", txnReq.Data.AccountNonce, sv.pbftSv.Nd.ID)

	return &pb.GenericResp{Msg: fmt.Sprintf("Transaction request %d received in node %d", txnReq.Data.AccountNonce, sv.pbftSv.Nd.ID)}, nil
}

// RegisterPbftGrpcListener listens to client for new transaction requests on grpcPort
func RegisterPbftGrpcListener(grpcPort int, sv *Server) {
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
func BuildServer(cfg *Config, me int) *Server {
	sv := &Server{}
	sv.IP = cfg.Network.IPList[me]
	sv.Port = cfg.Network.Ports[me]
	sv.Cfg = cfg
	sv.Nd = Make(cfg, me, cfg.Network.Ports[me], 0)

	RegisterPbftGrpcListener(cfg.Network.GrpcPorts[me], sv)

	go sv.Start() // in case the server has some initial logic
	return sv
}
