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
	"log"
	"os"
)

// const PORT_NUMBER = 40623
// const MAX_FAIL = 1
const OUTPUT_THRESHOLD = 1
const BASE_PORT = 40540

type Config struct {
	N          int
	IPList     []string
	Ports      []int
	HOSTS_FILE string
}

func (cfg *Config) Generate_keys() {
	N := 1000
	KD := "keys/"
	_, err := os.Stat(KD)
	if os.IsNotExist(err) {
		err := os.Mkdir(KD, 0777)
		if err != nil {
			log.Fatalln(err)
		}
	}

	write_new_keys(N)
	fmt.Printf("Generated %d keys in keys/ folder..\n", N)
}
