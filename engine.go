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
	"fmt"
	"os"
  "path"
	"./pbft-core"
	)

func (cfg *pbft.config) generate_keys(){
	N := 1000
	KD := "keys/"
	_, err := os.Stat(KD)
        if os.IsNotExist(err) {
                err := os.Mkdir(KD, 0777)
                if err != nil {
                        log.Fatalln(err)
                }
        }

	pbft.write_new_keys(N)
	fmt.Printf("Generated %d keys in keys/ folder..\n", N)
}

func main(){
	cfg := pbft.config{}
	HOSTS_FILE = path.Join(os.Getenv("HOME"), "hosts")

	cfg.IPList, cfg.Ports = getIPConfigs(pbft.HOSTS_FILE)
	cfg.N = len(cfg.IPList)
	cfg.generate_keys()
}
