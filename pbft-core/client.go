/*
The MIT License (MIT)

Copyright (c) 2018 TrueChain Foundation

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package pbft;

import (
	"net/rpc"
	"path"
	"os"
	"strconv"
	"fmt"
	"crypto/ecdsa"
)

// import "fmt"
// import "net"

type Client struct {
	IP string
	Port int
	Index int
	Me int
	Cfg *Config
	privKey ecdsa.PrivateKey
	peers []*rpc.Client  // directly contact Server.Nd
}

func (cl *Client) Start() {
	myPrint(1, "Firing up client executioner...\n")

}

func (cl *Client) NewRequest(msg string, timeStamp i) {
	//broadcast the request
	for i:=0; i<cl.Cfg.N; i++ {
		req := Request{RequestInner{cl.Cfg.N,0, 0, TYPE_REQU, msg, timeStamp, nil}, "", msgSignature{nil, nil}}  // the N-th party is the client
		req.inner.outer = &req
		req.addSig(&cl.privKey)
		arg := ProxyNewClientRequestArg{req, cl.Me}
		reply := ProxyNewClientRequestReply{}
		cl.peers[i].Go("Node.NewClientRequest", arg, &reply, nil)
	}
}

func BuildClient(cfg Config, IP string, Port int, me int) *Client {
	cl := &Client{}
	cl.IP = IP
	cl.Port = Port
	cl.Me = me
	cl.Cfg = &cfg
	peers := make([]*rpc.Client, cfg.N)
	for i:= 0; i<cfg.N; i++ {
		cl, err := rpc.Dial("tcp", cfg.IPList[i] + ":" + strconv.Itoa(cfg.Ports[i]))
		if err != nil {
			myPrint(3, "RPC error.\n")
		}
		peers[i] = cl
	}
	cl.peers = peers

	// TODO: prepare ecdsa private key for the client
	go cl.Start() // in case client has some initial logic
	return cl
}