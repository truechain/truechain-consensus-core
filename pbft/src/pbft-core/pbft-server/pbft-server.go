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
	"fmt"
	"log"
	"math/big"
	"net"

	"golang.org/x/net/context"

	"pbft-core"

	"google.golang.org/grpc"
	pb "pbft-core/fastchain"
)

const (
	port = 10000
)

// pbftserver defines the base properties of a pbft node server
type PbftServer struct {
	IP   string
	Port int
	Nd   *pbft.Node
	Cfg  *pbft.Config
	Out  chan pbft.ApplyMsg
	//TxnChan chan pb.Transaction
}

type fastChainServer struct {
	pbftSv *PbftServer
}

// Start - Initial server logic goes here
func (sv *PbftServer) Start() {
	pbft.MyPrint(1, "Firing up peer server...\n")
}

func (sv *fastChainServer) CheckLeader(context.Context, *pb.CheckLeaderReq) (*pb.CheckLeaderResp, error) {
	return &pb.CheckLeaderResp{Message: sv.pbftSv.Nd.Primary == sv.pbftSv.Nd.ID}, nil
}

func verifyTxnReq(req *pb.Request) error {
	// TODO: Transaction verification logic goes here
	fmt.Println("Verifying transaction request")
	return nil
}

func createInternalPbftReq(txnReq *pb.Request) pbft.Request {
	req := pbft.Request{}
	reqInner := pbft.RequestInner{}

	reqInner.ID = int(txnReq.Inner.Id)
	reqInner.Seq = int(txnReq.Inner.Seq)
	reqInner.View = int(txnReq.Inner.View)
	reqInner.Reqtype = int(txnReq.Inner.Type)
	reqInner.Msg = pbft.MsgType(txnReq.Inner.Msg[:])
	reqInner.Timestamp = txnReq.Inner.Timestamp

	sig := pbft.MsgSignature{}
	sig.R = big.NewInt(txnReq.Sig.R)
	sig.S = big.NewInt(txnReq.Sig.S)

	req.Inner = reqInner
	req.Sig = sig
	req.Dig = pbft.DigType(txnReq.Dig[:])

	return req
}

func (sv *fastChainServer) NewTxnRequest(ctx context.Context, txnReq *pb.Request) (*pb.GenericResp, error) {
	_ = verifyTxnReq(txnReq)

	req := createInternalPbftReq(txnReq)
	sv.pbftSv.Nd.NewClientRequest(req, req.Inner.ID)

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
	//sv.TxnChan = make(chan pb.Transaction)

	applyChan := make(chan pbft.ApplyMsg, cfg.NumQuest)

	go func(aC chan pbft.ApplyMsg) {
		for {
			c := <-aC
			pbft.MyPrint(4, "[0.0.0.0:%d] [%d] New Sequence Item: %v\n", sv.Port, me, c)
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
