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
	"path"
)

// MaxFail defines the number of faults to be tolerated
const MaxFail = 1

// OutputThreshold is an auxiliary to utils.MyPrint for verbosity
const OutputThreshold = 0

// BasePort is used as base value of ports defined for initializng PBFT servers
const BasePort = 40540

//GrpcBasePort is used as base value of ports defined for grpc communication between PBFT servers and clients
const GrpcBasePort = 10000

// Config is a generic struct used by each node to interact with connection pool details
type Config struct {
	N         int      // number of nodes to be launchedt
	KD        string   // key directory where pub/priva ECDSA keys are stored
	LD        string   // log directory
	IPList    []string // stores list of IP addresses belonging to BFT nodes
	Ports     []int    // stores list of Ports belonging to BFT nodes
	GrpcPorts []int    // stores list of ports serving grpc requests
	HostsFile string   // network config file, to read IP addresses from
	NumQuest  int      // NumQuest is the number of requests sent from client
	NumKeys   int      // NumKeys is the count of IP addresses (BFT nodes) participating
}

// GenerateKeysToFile generates ECDSA public-private keypairs to a folder
func (cfg *Config) GenerateKeysToFile(numKeys int) {
	// IdCount := 1000
	cfg.KD = path.Join(GetCWD(), "keys/")
	MakeDirIfNot(cfg.KD)
	WriteNewKeys(numKeys, cfg.KD)
	MyPrint(1, "Generated %d keypairs in %s folder..\n", numKeys, cfg.KD)
}
