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
	"path"
	"pbft"
	"strconv"
	"time"
)

func main() {

	cfg := pbft.Config{}
	cfg.HostsFile = path.Join(os.Getenv("HOME"), "hosts")
	cfg.IPList, cfg.Ports = pbft.GetIPConfigs(cfg.HostsFile)
	fmt.Printf("Get IPList %v, Ports %v\n", cfg.IPList, cfg.Ports)
	cfg.NumKeys = len(cfg.IPList)
	cfg.N = cfg.NumKeys - 1 // we assume client count to be 1
	cfg.NumQuest = 100
	cfg.GenerateKeysToFile(cfg.NumKeys)

	svList := make([]*pbft.Server, cfg.N)
	for i := 0; i < cfg.N; i++ {
		fmt.Println(cfg.IPList[i], cfg.Ports[i], i)
		svList[i] = pbft.BuildServer(cfg, cfg.IPList[i], cfg.Ports[i], i)
	}
	for i := 0; i < cfg.N; i++ {
		<-svList[i].Nd.ListenReady
	}
	time.Sleep(1 * time.Second) // wait for the servers to accept incoming connections
	for i := 0; i < cfg.N; i++ {
		svList[i].Nd.SetupReady <- true // make them to dial each other's RPCs
	}
	fmt.Println("[!!!] Please allow the program to accept incoming connections if you are using Mac OS.")
	time.Sleep(1 * time.Second) // wait for the servers to accept incoming connections

	cl := pbft.BuildClient(cfg, cfg.IPList[cfg.N], cfg.Ports[cfg.N], 0)
	start := time.Now()
	for k := 0; k < cfg.NumQuest; k++ {
		cl.NewRequest("Request "+strconv.Itoa(k), int64(k))
	}
	fmt.Println("Finish sending the requests.")
	finish := make(chan bool)
	for i := 0; i < cfg.N; i++ {
		go func(ind int) {
			for {
				c := <-svList[ind].Out
				if c.Index == cfg.NumQuest {
					finish <- true
				}
			}

		}(i)
	}
	<-finish
	elapsed := time.Since(start)
	fmt.Println("Test finished. Time cost:", elapsed)
}
