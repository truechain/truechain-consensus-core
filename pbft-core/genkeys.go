/*
The MIT License (MIT)

Copyright (c) 2018 TrueChain Foundation

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package pbft

import (
	"os"
	"fmt"
	"log"
	"path"
	"io/ioutil"
	"crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
	"path/filepath"
	"encoding/gob"
	"bytes"
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

	for k:= 0; k < kcount; k++ {
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
