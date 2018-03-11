package pbft;

import "net"
import "fmt"
import "net/rpc"
import (
	"strconv"
	"net/http"
)

// TODO: implement send/receive messages with sockets

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

