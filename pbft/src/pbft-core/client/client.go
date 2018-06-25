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
	"fmt"
	"log"
	"path"
	"strconv"
	"time"

	"pbft-core"
	"pbft-core/pbft-server"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "pbft-core/fastchain"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

var (
	cfg    = pbft.Config{}
	svList []*pbftserver.PbftServer
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
	pbft.MyPrint(1, "Firing up client executioner...\n")

}

// LoadPbftClientConfig loads client configuration
func (cl *Client) LoadPbftClientConfig() {
	cl.IP = cfg.IPList[cfg.N]
	cl.Port = cfg.Ports[cfg.N]
	cl.Me = 0

	pemkeyFile := fmt.Sprintf("sign%v.pem", cfg.N)
	sk := pbft.FetchPrivateKey(path.Join(cfg.KD, pemkeyFile))
	fmt.Println("just fetched private key for Client")
	fmt.Println(sk)
	cl.privKey = sk
}

func addSig(txnData *pb.TxnData, privKey *ecdsa.PrivateKey) {
	//MyPrint(1, "adding signature.\n")
	/*gob.Register(&pb.Request_Inner{})
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(req.Inner)
	if err != nil {
		pbft.MyPrint(3, "%s err:%s", `failed to encode!\n`, err.Error())
		return
	}*/

	/*s := pbft.GetHash(string(txnData.Payload))
	pbft.MyPrint(1, "digest %s.\n", string(s))
	req.Dig = []byte(pbft.DigType(s))*/
	if privKey != nil {
		hashData := ethcrypto.Keccak256(txnData.Payload)
		txnData.Signature, _ = ethcrypto.Sign(hashData, privKey)
		/*sigr, sigs, err := ecdsa.Sign(rand.Reader, privKey, []byte(s))
		if err != nil {
			pbft.MyPrint(3, "%s", "Error signing.")
			return
		}*/
	}
}

// NewRequest takes in a message and timestamp as params for a new request from client
func (cl *Client) NewRequest(msg string, timeStamp int64) {
	//broadcast the request
	for i := 0; i < cfg.N; i++ {
		txnreq := &pb.Transaction{
			Data: &pb.TxnData{
				AccountNonce: 0,
				Price:        0,
				GasLimit:     0,
				Recipient:    []byte(""),
				Amount:       0,
				Payload:      []byte(msg),
				Hash:         []byte(""),
				Signature:    []byte(""),
			},
		}

		addSig(txnreq.Data, cl.privKey)
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", cfg.GrpcPorts[i]), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		c := pb.NewFastChainClient(conn)
		ctx := context.TODO()

		resp, err := c.NewTxnRequest(ctx, txnreq)
		if err != nil {
			log.Fatalf("could not send transaction request to pbft node: %v", err)
		}

		fmt.Printf("%s\n", resp.Msg)
		conn.Close()
		break
	}
}

func main() {
	cfg.LoadPbftSimConfig()
	cl := &Client{}
	cl.LoadPbftClientConfig()

	go cl.Start() // in case client has some initial logic

	start := time.Now()
	for k := 0; k < cfg.NumQuest; k++ {
		cl.NewRequest("Request "+strconv.Itoa(k), time.Now().Unix()) // A random string Request{1,2,3,4....}
	}

	fmt.Println("Finish sending the requests.")

	/*finish := make(chan bool)
	for i := 0; i < cfg.N; i++ {
		go func(ind int) {
			for {
				// place where channel data is extracted out of Node's channel context
				c := <-svList[ind].Out
				if c.Index == cfg.NumQuest {
					finish <- true
				}
			}

		}(i)
	}
	<-finish*/
	elapsed := time.Since(start)
	fmt.Println("Test finished. Time cost:", elapsed)
}
