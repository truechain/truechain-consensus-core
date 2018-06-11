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

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"

	"golang.org/x/net/context"

	"pbft-core"

	"google.golang.org/grpc"
	pb "pbft-core/fastchain"
)

var (
	port = flag.Int("port", 10000, "The server port")
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

// Start - Initial server logic goes here
func (sv *PbftServer) Start() {
	pbft.MyPrint(1, "Firing up peer server...\n")
}

func (sv *PbftServer) CheckLeader(context.Context, *pb.CheckLeaderReq) (*pb.CheckLeaderResp, error) {
	return &pb.CheckLeaderResp{Message: sv.Nd.Primary == sv.Nd.ID}, nil
}

func verifyTxn(txn *pb.Transaction) error {
	// TODO: Transaction verification logic goes here
	fmt.Println("Verifying transaction request")
	return nil
}

func (sv *PbftServer) NewTxnRequest(ctx context.Context, txn *pb.Transaction) (*pb.GenericResp, error) {
	_ = verifyTxn(txn)
	return &pb.GenericResp{Msg: "Transaction request received"}, nil
}

// BuildServer initiates the Server resource properties and listens to client's
// message requests as well as interacts with the channel
func BuildServer(cfg pbft.Config, IP string, Port int, me int) *PbftServer {
	sv := &PbftServer{}
	sv.IP = IP
	sv.Port = Port
	sv.Out = make(chan pbft.ApplyMsg)
	sv.Cfg = &cfg
	//sv.TxnChan = make(chan pb.Transaction)

	applyChan := make(chan pbft.ApplyMsg)

	go func(aC chan pbft.ApplyMsg) {
		for {
			c := <-aC
			pbft.MyPrint(4, "[0.0.0.0:%d] [%d] New Sequence Item: %v\n", sv.Port, me, c)
			sv.Out <- c
		}
	}(applyChan)
	sv.Nd = pbft.Make(cfg, me, Port, 0, applyChan, 100) // test 100 messages

	//go sv.Start() // in case the server has some initial logic
	return sv
}

func main() {
	cfg := pbft.Config{}
	cfg.HostsFile = path.Join(os.Getenv("HOME"), "hosts") // TODO: read from config.yaml in future.
	cfg.IPList, cfg.Ports = pbft.GetIPConfigs(cfg.HostsFile)
	cfg.NumKeys = len(cfg.IPList)
	cfg.N = cfg.NumKeys - 1 // we assume client count to be 1
	cfg.NumQuest = 100
	cfg.GenerateKeysToFile(cfg.NumKeys)

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterPbftServer(grpcServer, BuildServer(cfg, cfg.IPList[0], cfg.Ports[0], 0))
	grpcServer.Serve(lis)
}
