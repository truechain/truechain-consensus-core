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

package main

import (
	//"fmt"
	"os"
    "path"
	//"log"
	"./pbft-core"
	"strconv"
	"fmt"
)

const NUM_QUEST = 1000

func main(){

	cfg := pbft.Config{}
	cfg.HOSTS_FILE = path.Join(os.Getenv("HOME"), "hosts")
	cfg.IPList, cfg.Ports = pbft.GetIPConfigs(cfg.HOSTS_FILE)
	cfg.N = len(cfg.IPList) - 1  // we assume the number of client is 1
	cfg.Generate_keys()
	/////////////////

	svList := make([]*pbft.Server, cfg.N)
	for i:=0; i<cfg.N; i++ {
		svList[i] = pbft.BuildServer(cfg, cfg.IPList[i], cfg.Ports[i], i)
	}
	for i:=0; i<cfg.N; i++ {
		<- svList[i].Nd.ListenReady
	}
	for i:=0; i<cfg.N; i++ {
		svList[i].Nd.SetupConnections()
	}

	/////////////////
	cl := pbft.BuildClient(cfg, cfg.IPList[cfg.N], cfg.Ports[cfg.N], 0)
	for k:=0; k < NUM_QUEST; k++ {
		cl.NewRequest("Request " + strconv.Itoa(k), int64(k))
	}
	fmt.Println("Finish sending the requests.\n")
}
