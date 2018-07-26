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

	"golang.org/x/sys/unix"
	"trueconsensus/fastchain"
	"trueconsensus/logging"
)

var (
	cfg    = pbft.GetPbftConfig()
	svList []*pbft.Server
)

// StartPbftServers launches the setup with numq count of messages
func StartPbftServers() {
	svList = make([]*pbft.Server, cfg.N)
	for i := 0; i < cfg.N; i++ {
		fmt.Println(cfg.Network.IPList[i], cfg.Network.Ports[i], i)
		svList[i] = pbft.BuildServer(cfg, i)
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

	// initial the root logger
	logging.InitRootLoger()
	StartPbftServers()
	cfg.GenerateKeysToFile()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case sig := <-c:
			fmt.Printf("Got %s signal. Aborting...\n", sig)
			os.Exit(1)
		}
	}()

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
