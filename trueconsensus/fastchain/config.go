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
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"trueconsensus/common"
)

const (
	configFile               = "/.truechain_config.yml"
	trueChainConsensusConfig = "TRUECHAIN_CONSENSUS_CONFIGURATION"
)

// // MaxFail defines the number of faults to be tolerated
// const MaxFail = 1

// // // OutputThreshold is an auxiliary to utils.MyPrint for verbosity
// // const OutputThreshold = 0

// // BasePort is used as base value of ports defined for initializng PBFT servers
// const BasePort = 40540
// //GrpcBasePort is used as base value of ports defined for grpc communication between PBFT servers and clients
// const GrpcBasePort = 10000

// Config  configuration for pbft
type Config struct {
	// general
	BasePort          int   `yaml:"basePort"`
	GrpcBasePort      int   `yaml:"grpcBasePort"`
	MaxFail           int   `yaml:"maxFail"`
	TorSocksPortRange []int `yaml:"torSocksPortRange"`

	// Node - generic struct used by each node to interact with connection pool details
	Node struct {
		N         int      // number of nodes to be launchedt
		KD        string   // key directory where pub/priva ECDSA keys are stored
		LD        string   // log directory
		IPList    []string // stores list of IP addresses belonging to BFT nodes
		Ports     []int    // stores list of Ports belonging to BFT nodes
		GrpcPorts []int    // stores list of ports serving grpc requests
		HostsFile string   // network config file, to read IP addresses from
		NumQuest  int      // NumQuest is the number of requests sent from client
		NumKeys   int      // NumKeys is the count of IP addresses (BFT nodes) participating
		Blocksize int      // Blocksize specifies the number of transactions per block
	}

	// bftCommittee config
	BftCommittee struct {
		ActualDelta int `yaml:"actualDelta"`
		Csize       int `yaml:"csize"`
		Delta       int `yaml:"delta"`
		Lambda      int `yaml:"lambda"`
		Tbft        int `yaml:"tbft"`
		Th          int `yaml:"th"`
		Timeout     int `yaml:"timeout"`
		// TODO: add []chain
		// Chain []struct {
		// }
	}

	// slow chain
	Slowchain struct {
		Csize int `yaml:"csize"`
	}

	// testbed config
	TestbedConfig struct {
		ClientID     int `yaml:"clientID"`
		InitServerID int `yaml:"initServerID"`
		Requests     struct {
			Max       int `yaml:"max"`
			BatchSize int `yaml:"batchSize"`
		}
		ThreadingEnabled bool `yaml:"threadingEnabled"`
		Total            int  `yaml:"total"`
	}
}

func getFilePath() string {
	path := os.Getenv(trueChainConsensusConfig)
	if path == "" {
		path = os.Getenv("HOME") + configFile
	}
	return path
}

/*
// LoadLogisticsIni loads trueconsensus/config/logistcs_bft.cfg
func LoadLogisticsIni() (configData *ini.File, err error) {
     path := path2.Join(os.Getenv("PWD"), "cfg.ini")
    configData, err = ini.Load(path)
    if err != nil {
    	fmt.Printf("Error reading ini file: %v", err)
    }
    return configData, nil
}
*/

// LoadTunablesConfig loads trueconsensus/config/tunables_bft.yaml
func LoadTunablesConfig() (*Config, error) {

	filePath := getFilePath()
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Unable to read config file. Error:%+v \n", err)
		return nil, err
	}

	/*
		cfgData, err := LoadConfig()
		common.CheckErr(err)
		if cfgData != nil {

			nd.cfg.LD = cfgData.Section("log").Key("log_folder").String()
		} else {
			nd.cfg.LD = path.Join(GetCWD(), "logs/")
		}
	*/

	config := &Config{}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Printf("Unable to Unmarshal config file. Error:%+v", err)
		return nil, err
	}

	err = validateConfig(config)
	if err != nil {
		log.Printf("Error: Validate config file values failed. Error: %+v", err)
		return nil, err
	}
	log.Printf("Info: initial the configuration for pbft \n")
	return config, nil
}

func validateConfig(config *Config) error {
	return nil
}

// GetIPConfigs loads all the IPs from the ~/hosts files
func GetIPConfigs(s string) ([]string, []int, []int) {
	// s: config file path
	common.MyPrint(1, "Loading IP configs...\n")
	contentB, err := ioutil.ReadFile(s)
	common.CheckErr(err)
	content := string(contentB)
	lst := strings.Fields(content)
	ports := make([]int, 0)
	grpcports := make([]int, 0)
	for k := range lst {
		ports = append(ports, BasePort+k)
		grpcports = append(grpcports, GrpcBasePort+k)
	}
	return lst, ports, grpcports
}

// GenerateKeysToFile generates ECDSA public-private keypairs to a folder
func (cfg *Config) GenerateKeysToFile(numKeys int) {
	MakeDirIfNot(cfg.KD)
	WriteNewKeys(numKeys, cfg.KD)

	MyPrint(1, "Generated %d keypairs in %s folder..\n", numKeys, cfg.KD)
}

// LoadPbftSimConfig loads configuration for running PBFT simulation
func (cfg *Config) LoadPbftSimConfig() {
	cfg.HostsFile = path.Join(os.Getenv("HOME"), "hosts") // TODO: read from config.yaml in future.
	cfg.IPList, cfg.Ports, cfg.GrpcPorts = GetIPConfigs(cfg.HostsFile)
	cfg.NumKeys = len(cfg.IPList)
	cfg.N = cfg.NumKeys - 1 // we assume client count to be 1
	// Load this from commandline/set default in client.go
	// cfg.NumQuest = 100
	cfg.Blocksize = 10 // This is hardcoded to 10 for now
	cfg.KD = path.Join(GetCWD(), "keys/")
}

// GetPbftConfig returns the basic PBFT configuration used for simulation
func GetPbftConfig() *Config {
	cfg := &Config{}
	cfg.LoadPbftSimConfig()

	return cfg

}
