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
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

func GetCWD() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func write_new_keys(kcount int) {
	gob.Register(&ecdsa.PrivateKey{})

	for k := 0; k < kcount; k++ {
		sk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		filename := fmt.Sprintf("sign%v.dat", k)
		// FIXME: gob (??)
		buf := bytes.Buffer{}
		e := gob.NewEncoder(&buf)
		e.Encode(sk)
		err := ioutil.WriteFile(path.Join(GetCWD(), "keys/", filename), buf.Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}
