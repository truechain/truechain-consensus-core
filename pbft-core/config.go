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
	// "fmt"
	// "log"
	// "os"
	"path"
	"crypto/ecdsa"
	"crypto/elliptic"
	//"fmt"
	//"bytes"
	//"encoding/gob"
	//"io/ioutil"
	//"log"
	"crypto/rand"
)

// const PORT_NUMBER = 40623
// const MAX_FAIL = 1
const OUTPUT_THRESHOLD = 0
const BASE_PORT = 40540

const NUM_KEYS = 10

type Config struct {
	N          int
	KD				 string
	LD				 string
	IPList     []string
	Ports      []int
	HOSTS_FILE string
	Keys 		[]*ecdsa.PrivateKey
}

func (cfg *Config) GenerateKeys() {
	cfg.Keys = make([]*ecdsa.PrivateKey, NUM_KEYS)
	for k := 0; k < NUM_KEYS; k++ {
		sk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		cfg.Keys[k] = sk
	}
}

func (cfg *Config) GenerateKeysToFile() {
	IdCount := 1000
	cfg.KD = path.Join(GetCWD(), "keys/")
	MakeDirIfNot(cfg.KD)
	WriteNewKeys(IdCount, cfg.KD)
	MyPrint(1, "Generated %d keys in %s folder..\n", IdCount, cfg.KD)
}
