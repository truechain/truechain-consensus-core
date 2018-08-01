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
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"path"
	"strconv"
	"time"

	"trueconsensus/common"
	"trueconsensus/fastchain"

	"context"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	pb "trueconsensus/fastchain/proto"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

var (
	cfg    = pbft.GetPbftConfig()
	svList []*pbft.Server
	cl     = Client{}
)

// Client makes queries to what it believes to be the primary replica.
// Below defines the major properties of a client resource
type Client struct {
	IP      string
	Port    int
	Index   int
	Me      int
	privKey *ecdsa.PrivateKey
}

// Start is a notifier of client's init state
func (cl *Client) Start() {
	common.MyPrint(1, "Firing up client executioner...\n")

}

// LoadPbftClientConfig loads client configuration
func (cl *Client) LoadPbftClientConfig() {
	cl.IP = cfg.Network.IPList[cfg.Network.N]
	cl.Port = cfg.Network.Ports[cfg.Network.N]
	cl.Me = 0

	pemkeyFile := fmt.Sprintf("sign%v.pem", cfg.Network.N)
	sk, _ := ethcrypto.LoadECDSA(path.Join(cfg.Logistics.KD, pemkeyFile))
	fmt.Println("just fetched private key for Client")
	fmt.Println(sk)
	cl.privKey = sk
}

func (cl *Client) addSig(txnData *pb.TxnData) {
	if cl.privKey != nil {
		data, _ := proto.Marshal(txnData)
		hashData := ethcrypto.Keccak256(data)
		txnData.Hash = hashData
		sig, _ := ethcrypto.Sign(hashData, cl.privKey)
		txnData.Signature = sig
	}
}

// NewRequest takes in a message and timestamp as params for a new request from client
func (cl *Client) NewRequest(msg string, k int, timeStamp int64) {
	//broadcast the request
	for i := 0; i < cfg.Network.N; i++ {
		txnreq := &pb.Transaction{
			Data: &pb.TxnData{
				AccountNonce: uint64(k),
				Price:        int64(k),
				GasLimit:     0,
				Recipient:    []byte(""),
				Amount:       0,
				Payload:      []byte(msg),
			},
		}

		cl.addSig(txnreq.Data)
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", cfg.Network.GrpcPorts[i]), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		c := pb.NewFastChainClient(conn)
		ctx := context.TODO()

		resp, err := c.NewTxnRequest(ctx, txnreq)
		if err != nil {
			log.Fatalf("Transaction request to PBFT node failed: %v", err)
		}

		fmt.Printf("%s\n", resp.Msg)
		conn.Close()
	}
}

func main() {
	nReq := flag.Int("numquest", 10, "number of requests")
	flag.Parse()
	cfg.Network.NumQuest = *nReq
	log.Printf("REQUESTS count - %v", cfg.Network.NumQuest)

	cl := &Client{}
	cl.LoadPbftClientConfig()

	go cl.Start() // in case client has some initial logic

	start := time.Now()
	for k := 0; k < cfg.Network.NumQuest; k++ {
		cl.NewRequest("Request "+strconv.Itoa(k), k, time.Now().Unix()) // Transaction request where nonce = gasPrice = k
	}

	fmt.Println("Finish sending the requests.")

	elapsed := time.Since(start)
	fmt.Println("Test finished. Time cost:", elapsed)
}
