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
	"os"
	"os/signal"
	"time"

	"pbft-core"
	"pbft-core/pbft-server"

	"golang.org/x/sys/unix"
)

var (
	cfg    = pbft.Config{}
	svList []*pbftserver.PbftServer
)

// StartPbftServers starts PBFT servers from config information
func StartPbftServers() {
	svList = make([]*pbftserver.PbftServer, cfg.N)
	for i := 0; i < cfg.N; i++ {
		fmt.Println(cfg.IPList[i], cfg.Ports[i], i)
		svList[i] = pbftserver.BuildServer(cfg, cfg.IPList[i], cfg.Ports[i], cfg.GrpcPorts[i], i)
	}

	for i := 0; i < cfg.N; i++ {
		<-svList[i].Nd.ListenReady
	}

	time.Sleep(1 * time.Second) // wait for the servers to accept incoming connections
	for i := 0; i < cfg.N; i++ {
		svList[i].Nd.SetupReady <- true // make them to dial each other's RPCs
	}

	//fmt.Println("[!!!] Please allow the program to accept incoming connections if you are using Mac OS.")
	time.Sleep(1 * time.Second) // wait for the servers to accept incoming connections
}

func main() {
	cfg.LoadPbftSimConfig()
	StartPbftServers()

	finish := make(chan bool)
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
	<-finish

	// Use the main goroutine as signal handling loop
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh)
	for s := range sigCh {
		switch s {
		case unix.SIGTERM:
			fallthrough
		case unix.SIGINT:
			return
		default:
			continue
		}
	}
}
