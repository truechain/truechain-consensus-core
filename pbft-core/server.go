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
/*"path"
"os"
"net/rpc"
"strconv"*/
//"fmt"
)

// import "fmt"

type Server struct {
	IP   string
	Port int
	Nd   *Node
	Cfg  *Config
	Out  chan ApplyMsg
}

func (sv *Server) Start() {
	MyPrint(1, "Firing up peer server...\n")
}

func BuildServer(cfg Config, IP string, Port int, me int) *Server {
	sv := &Server{}
	sv.IP = IP
	sv.Port = Port
	sv.Out = make(chan ApplyMsg)
	sv.Cfg = &cfg
	applyChan := make(chan ApplyMsg)
	go func(aC chan ApplyMsg) {
		for {
			c := <-aC
			MyPrint(4, "[0.0.0.0:%d] [%d] New Sequence Item: %v\n", sv.Port, me, c)
			sv.Out <- c
		}
	}(applyChan)
	sv.Nd = Make(cfg, me, Port, 0, applyChan, 100) // test 100 messages

	go sv.Start() // in case the server has some initial logic
	return sv
}
