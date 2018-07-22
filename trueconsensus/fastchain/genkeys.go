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
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"trueconsensus/common"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// GetCWD provides the current work dir's absolute filepath,
// where CWD is the one where truechain-engine binary shots were was fired from.
func GetCWD() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

// WriteNewKeys generates kcount # of ECDSA public-private key pairs and writes them
// into a folder kdir.
func WriteNewKeys(kcount int, kdir string) {
	for k := 0; k < kcount; k++ {
		privateKey, _ := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
		publicKey := &privateKey.PublicKey

		pemEncoded := hex.EncodeToString(ethcrypto.FromECDSA(privateKey))
		pemEncodedPub := hex.EncodeToString(ethcrypto.FromECDSAPub(publicKey))

		pemkeyFname := fmt.Sprintf("sign%v.pem", k)
		err1 := ioutil.WriteFile(path.Join(kdir, pemkeyFname), []byte(pemEncoded), 0600)
		common.CheckErr(err1)
		pubkeyFname := fmt.Sprintf("sign%v.pub", k)
		err2 := ioutil.WriteFile(path.Join(kdir, pubkeyFname), []byte(pemEncodedPub), 0644)
		common.CheckErr(err2)
	}
}
