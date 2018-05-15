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

package tests

import (
        "testing"
)

type Config struct {
     Keys       []*ecdsa.PrivateKey // in-memory map that stores pub/priv key, for mock tests
}

// GenerateKeys generate 'NumKeys' amount of keys to a hashmap
// This is for mock test purposes
func (cfg *Config) MockGenKeys(){
	Keys := make([]*ecdsa.PrivateKey, NumKeys)
	for k := 0; k < NumKeys; k++ {
		sk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		Keys[k] = sk
	}
}

func TestKeyGen(t *testing.T) {
        t.Parallel()
	// TODO
}

