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
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"trueconsensus/common"
)

// MakeTransportations is a placeholder. All network related functions
// should be defined here. TODO: implement send/receive messages with RPC
func (cfg *Config) MakeTransportations(index int) []*rpc.Client {
	// index: the index if the server itself
	clientList := make([]*rpc.Client, 0)
	// serve RPC server
	rpc.Register(&Node{})
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Itoa(cfg.Network.Ports[index]))
	common.CheckErr(e)
	go http.Serve(l, nil)
	for ind, val := range cfg.Network.IPList {
		client, e := rpc.DialHTTP("tcp", val+":"+strconv.Itoa(cfg.Network.Ports[ind]))
		common.CheckErr(e)
		clientList = append(clientList, client)
	}
	return clientList
}
