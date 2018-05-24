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

import "net"

// import "fmt"
import "net/rpc"
import (
	"net/http"
	"strconv"
)

// MakeTransportations - TODO: implement send/receive messages with RPC
func MakeTransportations(index int) []*rpc.Client {
	// index: the index if the server itself
	clientList := make([]*rpc.Client, 0)
	lst, ports := GetIPConfigs("ipconfig")
	// serve RPC server
	rpc.Register(&Node{})
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Itoa(ports[index]))
	CheckErr(e)
	go http.Serve(l, nil)
	for ind, val := range lst {
		client, e := rpc.DialHTTP("tcp", val+":"+strconv.Itoa(ports[ind]))
		CheckErr(e)
		clientList = append(clientList, client)
	}
	return clientList
}
