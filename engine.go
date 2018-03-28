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
	//"fmt"
	"os"
	"path"
	//"log"
	"fmt"
	"strconv"

	"./pbft-core"
	"time"
)

const NUM_QUEST = 1000

func main() {

	cfg := pbft.Config{}
	cfg.HOSTS_FILE = path.Join(os.Getenv("HOME"), "hosts")
	cfg.IPList, cfg.Ports = pbft.GetIPConfigs(cfg.HOSTS_FILE)
	cfg.N = len(cfg.IPList) - 1 // we assume the number of client is 1
	cfg.GenerateKeys()
	/////////////////

	svList := make([]*pbft.Server, cfg.N)
	for i := 0; i < cfg.N; i++ {
		svList[i] = pbft.BuildServer(cfg, cfg.IPList[i], cfg.Ports[i], i)
	}
	for i := 0; i < cfg.N; i++ {
		<-svList[i].Nd.ListenReady
	}
	for i := 0; i < cfg.N; i++ {
		svList[i].Nd.SetupReady <- true // make them to dial each other's RPCs
	}
	fmt.Println("\n !!! Please permit the program to accept incoming connections if you are using Mac OS.")
	time.Sleep(5 * time.Second)  // wait for the servers to accept incoming connections
	/////////////////
	cl := pbft.BuildClient(cfg, cfg.IPList[cfg.N], cfg.Ports[cfg.N], 0)
	for k := 0; k < NUM_QUEST; k++ {
		cl.NewRequest("Request "+strconv.Itoa(k), int64(k))
	}
	fmt.Println("Finish sending the requests.")
}
