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
	// "bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	// "encoding/pem"
	// "crypto/x509"
	// "encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
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

// func encode(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) ([]byte, []byte) {
// 		x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
// 		pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
// 		x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
// 		pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
//     return []byte(pemEncoded), []byte(pemEncodedPub)
// }

// WriteNewKeys generates kcount # of ECDSA public-private key pairs and writes them
// into a folder kdir.
func WriteNewKeys(kcount int, kdir string) {
	for k := 0; k < kcount; k++ {
		privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		publicKey := &privateKey.PublicKey
		pemEncoded, pemEncodedPub := EncodeECDSAKeys(privateKey, publicKey)

		pemkeyFname := fmt.Sprintf("sign%v.pem", k)
		err1 := ioutil.WriteFile(path.Join(kdir, pemkeyFname), pemEncoded, 0644)
		CheckErr(err1)
		pubkeyFname := fmt.Sprintf("sign%v.pub", k)
		err2 := ioutil.WriteFile(path.Join(kdir, pubkeyFname), pemEncodedPub, 0644)
		CheckErr(err2)
	}
}
