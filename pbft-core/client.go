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
	"net/rpc"
	//"path"
	//"os"
	"strconv"
	//"fmt"
	//"bytes"
	"crypto/ecdsa"
	//"encoding/gob"
	"fmt"
	//"io/ioutil"
	"path"
)

// import "fmt"
// import "net"

type Client struct {
	IP      string
	Port    int
	Index   int
	Me      int
	Cfg     *Config
	privKey *ecdsa.PrivateKey
	peers   []*rpc.Client // directly contact Server.Nd
}

func (cl *Client) Start() {
	MyPrint(1, "Firing up client executioner...\n")

}

func (cl *Client) NewRequest(msg string, timeStamp int64) {
	//broadcast the request
	for i := 0; i < cl.Cfg.N; i++ {
		//req := Request{RequestInner{cl.Cfg.N,0, 0, TYPE_REQU, MsgType(msg), timeStamp, nil}, "", msgSignature{nil, nil}}  // the N-th party is the client
		req := Request{RequestInner{cl.Cfg.N, 0, 0, TYPE_REQU, MsgType(msg), timeStamp}, "", MsgSignature{nil, nil}} // the N-th party is the client
		//req.inner.outer = &req
		req.addSig(cl.privKey)
		arg := ProxyNewClientRequestArg{req, cl.Me}
		reply := ProxyNewClientRequestReply{}
		cl.peers[i].Go("Node.ProxyNewClientRequest", arg, &reply, nil) // try synchronize
	}
}

func BuildClient(cfg Config, IP string, Port int, me int) *Client {
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
	// sk := cfg.Keys[cfg.N]
	fmt.Println("just fetched private key for Client")
	fmt.Println(sk)
	cl.privKey = sk
	// TODO: prepare ecdsa private key for the client
	go cl.Start() // in case client has some initial logic
	return cl
}
