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
	"fmt"
	"github.com/go-ini/ini"
	path2 "path"
	"os"
)

// MaxFail defines the number of faults to be tolerated
const MaxFail = 1

// OutputThreshold is an auxiliary to utils.MyPrint for verbosity
const OutputThreshold = 0

// BasePort is used as base value of ports defined for initializng PBFT servers
const BasePort = 40540

// Config is a generic struct used by each node to interact with connection pool details
type Config struct {
	N         int      // number of nodes to be launchedt
	KD        string   // key directory where pub/priva ECDSA keys are stored
	LD        string   // log directory
	IPList    []string // stores list of IP addresses belonging to BFT nodes
	Ports     []int    // stores list of Ports belonging to BFT nodes
	HostsFile string   // network config file, to read IP addresses from
	NumQuest  int      // NumQuest is the number of requests sent from client
	NumKeys   int      // NumKeys is the count of IP addresses (BFT nodes) participating
}

// LoadConfig ini file
func LoadConfig() (configData *ini.File, err error) {
	path := path2.Join(os.Getenv("PWD"), "cfg.ini")
	configData, err = ini.Load(path)
	if err != nil {
		fmt.Printf("Error reading ini file: %v", err)
	}
	return configData, nil
}

// GenerateKeysToFile generates ECDSA public-private keypairs to a folder
func (cfg *Config) GenerateKeysToFile(numKeys int) {
	cfgData, err := LoadConfig()
	CheckErr(err)
	if cfgData != nil {
		cfg.KD = cfgData.Section("general").Key("pem_path").String()
	} else {
		cfg.KD = path2.Join(GetCWD(), "keys/")
	}
	// IdCount := 1000
	MakeDirIfNot(cfg.KD)
	WriteNewKeys(numKeys, cfg.KD)
	MyPrint(1, "Generated %d keypairs in %s folder..\n", numKeys, cfg.KD)
}
