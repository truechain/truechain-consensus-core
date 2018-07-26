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
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"trueconsensus/common"
)

const (
	tunablesConfigEnv = "TRUE_TUNABLES_CONF"
	generalConfigEnv  = "TRUE_GENERAL_CONF"
	peerNetworkEnv    = "TRUE_NETWORK_CONF"
	SimulatedEnv      = "TRUE_SIMULATION"
)

type Testbed struct {
	Total            int  `yaml:"total"`
	ClientID         int  `yaml:"client_id"`
	InitServerID     int  `yaml:"server_id_init"`
	ThreadingEnabled bool `yaml:"threading_enabled"`
	MaxRetries       int  `yaml:"max_retries"`
	MaxRequests      int  `yaml:"max_requests"`
	BatchSize        int  `yaml:"batch_size"`
}

type Slowchain struct {
	Csize int `yaml:"csize"`
}

// bftCommittee config
type BftCommittee struct {
	ActualDelta int `yaml:"actual_delta"`
	Delta       int `yaml:"delta"`
	Lambda      int `yaml:"lambda"`
	Tbft        int `yaml:"tbft"`
	Th          int `yaml:"th"`
	Timeout     int `yaml:"timeout"`
	Blocksize   int // Blocksize specifies the number of transactions per block
	// TODO: add []chain
	// Chain []struct {
	// }
}

type General struct {
	MaxFail         int `yaml:"max_fail"`
	BasePort        int `yaml:"rpc_base_port"`
	OutputThreshold int `yaml:"output_threshold"`
	MaxLogSize      int `yaml:"max_log_size"`
	GrpcBasePort    int `yaml:"grpc_base_port"`
}

type Tunables struct {
	// struct tags, to add metadata to a struct's fields
	// testbed config
	Testbed `yaml:"testbed"`
	// slow chain
	Slowchain    `yaml:"slowchain"`
	General      `yaml:"general"`
	BftCommittee `yaml:"bft_committee"`
}

type Logistics struct {
	LedgerLoc string
	LD        string
	ServerLog string
	ClientLog string
	KD        string // key directory where pub/priva ECDSA keys are stored

}

// Network - generic struct used by each node to interact with connection pool details
type Network struct {
	N         int      // number of nodes to be launchedt
	IPList    []string // stores list of IP addresses belonging to BFT nodes
	Ports     []int    // stores list of Ports belonging to BFT nodes
	GrpcPorts []int    // stores list of ports serving grpc requests
	NumQuest  int      // NumQuest is the number of requests sent from client
	NumKeys   int      // NumKeys is the count of IP addresses (BFT nodes) participating
	HostsFile string   // contains network addresses for server/client/peers
}

// Config - configuration for pbft.Config
// Note: struct fields must be public in order for
// unmarshal to correctly populate the data.
type Config struct {
	Tunables
	Logistics
	Network
}

func DefaultTunables() *Tunables {
	return &Tunables{
		Testbed: Testbed{
			ClientID:         5,
			MaxRequests:      10,
			ThreadingEnabled: true,
		},
		// slow chain
		Slowchain: Slowchain{
			Csize: 10,
		},
		General: General{
			BasePort:   40450,
			MaxLogSize: 1023303,
			MaxFail:    1,
		},
	}
}

// LoadLogisticsCfg loads the .cfg file
func LoadLogisticsCfg() (*ini.File, error) {
	path := os.Getenv(generalConfigEnv)
	if path == "" {
		path = "/etc/truechain/logistics_bft.cfg"
	}
	configData, err := ini.Load(path)
	if err != nil {
		fmt.Printf("Error reading ini file: %v", err)
		return configData, err
	}
	return configData, nil
}

// LoadTunablesConfig loads the .yaml file
func (cfg *Config) LoadTunablesConfig() error {

	path := os.Getenv(tunablesConfigEnv)
	if path == "" {
		path = "/etc/truechain/tunables_bft.yaml"
	}

	yamlFile, err := ioutil.ReadFile(path)

	if err != nil {
		log.Printf("Unable to read config file. Error:%+v \n", err)
		return err
	}

	tunables := Tunables{}
	err = yaml.Unmarshal(yamlFile, &tunables)
	if err != nil {
		log.Printf("Unable to Unmarshal config file. Error:%+v", err)
		return err
	}

	if err != nil {
		log.Printf("Error: Validate config file values failed. Error: %+v", err)
		return err
	}

	cfg.Tunables = tunables
	return nil
}

// GenerateKeysToFile generates ECDSA public-private keypairs to a folder
func (cfg *Config) GenerateKeysToFile() {
	common.MakeDirIfNot(cfg.Logistics.KD)
	WriteNewKeys(cfg.Network.NumKeys, cfg.Logistics.KD)

	common.MyPrint(1, "Generated %d keypairs in %s folder..\n", cfg.Network.NumKeys, cfg.Logistics.KD)
}

// GetIPConfigs loads all the IPs from the ~/hosts files
func (cfg *Config) GetIPConfigs() {
	// s: config file path
	common.MyPrint(1, "Loading IP configs...\n")
	contentB, err := ioutil.ReadFile(cfg.Network.HostsFile)
	common.CheckErr(err)
	content := string(contentB)
	cfg.Network.IPList = strings.Fields(content)
	cfg.Network.Ports = make([]int, 0)
	cfg.Network.GrpcPorts = make([]int, 0)
	for k := range cfg.Network.IPList {
		cfg.Network.Ports = append(cfg.Network.Ports, cfg.Tunables.General.BasePort+k)
		cfg.Network.GrpcPorts = append(cfg.Network.GrpcPorts, cfg.Tunables.General.GrpcBasePort+k)
	}
}

// ValidateConfig checks for aberrance and loads struct Config{} with tunables and logistics vars
func (cfg *Config) ValidateConfig(cfgData *ini.File) error {
	// TODO: refer to validate yaml from browbeat
	cfg.Network.HostsFile = os.Getenv(peerNetworkEnv)
	if cfg.Network.HostsFile == "" {
		cfg.Network.HostsFile = "/etc/truechain/hosts"
	}
	cfg.GetIPConfigs()
	cfg.Network.NumKeys = len(cfg.Network.IPList)
	cfg.Network.N = cfg.Network.NumKeys - 1 // we assume client count to be 1
	cfg.BftCommittee.Blocksize = 10         // This is hardcoded to 10 for now
	return nil
}

func CheckErr(err error) {
	if err != nil {
		log.Printf("Error: Validate config file values failed. Error: %+v", err)
	}
}

// GetPbftConfig returns the basic PBFT configuration used for simulation
func GetPbftConfig() *Config {
	cfg := &Config{}
	err := cfg.LoadTunablesConfig()
	CheckErr(err)
	cfgData, err := LoadLogisticsCfg()
	CheckErr(err)

	// CheckErr(err)
	if cfgData != nil {
		cfg.Logistics.KD = cfgData.Section("general").Key("pem_keystore_path").String()
		cfg.Logistics.LedgerLoc = cfgData.Section("node").Key("ledger_location").String()
		cfg.Logistics.LD = cfgData.Section("log").Key("root_folder").String()
		cfg.Logistics.ServerLog = cfgData.Section("log").Key("server_logfile").String()
		cfg.Logistics.ClientLog = cfgData.Section("log").Key("client_logfile").String()
		logSize, _ := cfgData.Section("log").Key("max_log_size").Int()
		cfg.General.MaxLogSize = logSize
		log.Printf("Loaded logistics configuration.")
	} else {
		log.Printf("!! Unable to find required sections.")
		cfg.Logistics.KD = path.Join("./keys/")
		cfg.Logistics.LedgerLoc = "/var/log/truechain"
		cfg.Logistics.LD = "/var/log/truechain"
		cfg.Logistics.ServerLog = "engine.log"
		cfg.Logistics.ClientLog = "client.log"
		cfg.General.MaxLogSize = 4194304
	}
	err = cfg.ValidateConfig(cfgData)
	CheckErr(err)

	log.Println("---> using following configurations for project:")
	// log.Printf("---> using following configurations:\n%+v\n\n", cfg)
	yamlDebugInfo, _ := yaml.Marshal(&cfg)
	fmt.Printf("%+v\n", string(yamlDebugInfo))

	return cfg
}

// func main() {
// 	cfg := GetPbftConfig()
// 	fmt.Println(cfg)
// }
