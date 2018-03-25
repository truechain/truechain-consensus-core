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

import "net"
// import "fmt"
import "net/rpc"
import (
	"strconv"
	"net/http"
)

// TODO: implement send/receive messages with RPC

func MakeTransportations(index int) []*rpc.Client {
	// index: the index if the server itself
	clientList := make([]*rpc.Client, 0)
	lst, ports := getIPConfigs("ipconfig")
	// serve RPC server
	node := new(Node)
	rpc.Register(node)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":" + strconv.Itoa(ports[index]))
	checkErr(e)
	go http.Serve(l, nil)
	for ind, val := range lst {
		client, e := rpc.DialHTTP("tcp", val + ":" + strconv.Itoa(ports[ind]))
		checkErr(e)
		clientList = append(clientList, client)
	}
	return clientList
}
