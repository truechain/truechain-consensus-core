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
	"fmt"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "pbft-core/fastchain"
)

// Client makes queries to what it believes to be the primary replica.
// Below defines the major properties of a client resource
/*type Client struct {
	IP      string
	Port    int
	Index   int
	Me      int
	Cfg     *pbft.Config
	privKey *ecdsa.PrivateKey
	peers   []*rpc.Client // directly contact Server.Nd
}

// Start is a notifier of client's init state
func (cl *Client) Start() {
	MyPrint(1, "Firing up client executioner...\n")

}

// NewRequest takes in a message and timestamp as params for a new request from client
func (cl *Client) NewRequest(msg string, timeStamp int64) {
	//broadcast the request
	for i := 0; i < cl.Cfg.N; i++ {
		//req := Request{RequestInner{cl.Cfg.N,0, 0, typeRequest, MsgType(msg), timeStamp, nil}, "", msgSignature{nil, nil}}  // the N-th party is the client
		req := Request{RequestInner{cl.Cfg.N, 0, 0, typeRequest, MsgType(msg), timeStamp}, "", MsgSignature{nil, nil}} // the N-th party is the client
		//req.inner.outer = &req
		req.addSig(cl.privKey)
		arg := ProxyNewClientRequestArg{req, cl.Me}
		reply := ProxyNewClientRequestReply{}
		cl.peers[i].Go("Node.ProxyNewClientRequest", arg, &reply, nil) // try synchronize
	}
}*/

// BuildClient dials up connections to tease and test hot bft nodes
/*func BuildClient(cfg Config, IP string, Port int, me int) *Client {
	cl := &Client{}
	cl.IP = IP
	cl.Port = Port
	cl.Me = me
	cl.Cfg = &cfg
	peers := make([]*rpc.Client, cfg.N)
	for i := 0; i < cfg.N; i++ {
		cl, err := rpc.Dial("tcp", cfg.IPList[i]+":"+strconv.Itoa(cfg.Ports[i]))
		if err != nil {
			MyPrint(3, "RPC error.\n")
		}
		peers[i] = cl
	}
	cl.peers = peers
	pemkeyFile := fmt.Sprintf("sign%v.pem", cfg.N)
	sk := FetchPrivateKey(path.Join(cfg.KD, pemkeyFile))
	fmt.Println("just fetched private key for Client")
	fmt.Println(sk)
	cl.privKey = sk
	go cl.Start() // in case client has some initial logic
	return cl
}*/

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:10000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPbftClient(conn)

	// Contact the server and print out its response.
	ctx := context.TODO()

	checkLeaderResp, err := c.CheckLeader(ctx, &pb.CheckLeaderReq{})
	if err != nil {
		log.Fatalf("could not check if node is leader: %v", err)
	}

	fmt.Printf("%d", checkLeaderResp.Message)

	txnData := &pb.TxnData{}
	txnResp, err := c.NewTxnRequest(ctx, &pb.Transaction{Data: txnData})
	if err != nil {
		log.Fatalf("could not send transaction request to pbft node: %v", err)
	}

	fmt.Printf("%d", txnResp.Msg)
}
