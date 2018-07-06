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

package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

const (
	configFile               = "/.truechain_config.yml"
	trueChainConsensusConfig = "TRUECHAIN_CONSENSUS_CONFIGURATION"
)

// Config  configuration for pbft
type Config struct {
	// general
	BasePort          int   `yaml:"basePort"`
	GrpcBasePort      int   `yaml:"grpcBasePort"`
	MaxFail           int   `yaml:"maxFail"`
	TorSocksPortRange []int `yaml:"torSocksPortRange"`

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

func initConfig() (*Config, error) {

	filePath := getFilePath()
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Unable to read config file. Error:%+v \n", err)
		return nil, err
	}

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
