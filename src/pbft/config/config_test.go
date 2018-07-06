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
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	filePath, err := filepath.Abs("/src/support/config.yaml")
	assert.Nil(t, err)
	os.Setenv(trueChainConsensusConfig, filePath)
	config, err := initConfig()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(config.TorSocksPortRange))
}
