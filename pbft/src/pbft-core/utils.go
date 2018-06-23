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
	"io/ioutil"

	"github.com/fatih/color"
	// "encoding/hex"
	// "crypto/rand"
	"crypto/x509"
	// "github.com/ethereum/go-ethereum/common/math"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	// "io"
)

// MyPrint provides customized colored output functionality
func MyPrint(t int, format string, args ...interface{}) {
	// t: log level
	// 		0: information
	//		1: emphasis
	//		2: warning
	//		3: error
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	deepPurple := color.New(color.FgGreen).SprintFunc()

	if t >= OutputThreshold {

		switch t {
		case 0: // info
			fmt.Printf("[ ]"+format, args...)
			break
		case 1: // emphasized
			fmt.Printf(blue("[.]")+format, args...)
			break
		case 2: // warning
			fmt.Printf(yellow("[!]")+format, args...)
			break
		case 3: // error
			fmt.Printf(red("[x]")+format, args...)
			break
		case 4: // magical shit
			fmt.Printf(deepPurple("[x]")+format, args...)
		}
	}
}

// CheckErr simply checks errors and panics, as a guilty pleasure.
func CheckErr(e error) {
	if e != nil {
		panic(e)
	}
}

//MakeDirIfNot handles dir creation operations
func MakeDirIfNot(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.Mkdir(dir, 0750)
		CheckErr(err)
	}
}

// FetchPublicKey reads and decodes a public key from file stored on disk
func FetchPublicKey(kpath string) *ecdsa.PublicKey {
	encodedKey, errRead := ioutil.ReadFile(kpath)
	CheckErr(errRead)
	blockPub, _ := pem.Decode([]byte(encodedKey))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)
	return publicKey
}

// FetchPrivateKey reads and decodes a private key from file stored on disk
func FetchPrivateKey(kpath string) *ecdsa.PrivateKey {
	encodedKey, errRead := ioutil.ReadFile(kpath)
	CheckErr(errRead)
	block, _ := pem.Decode([]byte(encodedKey))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)
	return privateKey
}

// EncodeECDSAKeys takes in ECDSA objects for key pairs and returns encoded []byte formats
func EncodeECDSAKeys(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) ([]byte, []byte) {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	return []byte(pemEncoded), []byte(pemEncodedPub)
}

// GetIPConfigs loads all the IPs from the ~/hosts files
func GetIPConfigs(s string) ([]string, []int, []int) {
	// s: config file path
	MyPrint(1, "Loading IP configs...\n")
	contentB, err := ioutil.ReadFile(s)
	CheckErr(err)
	content := string(contentB)
	lst := strings.Fields(content)
	ports := make([]int, 0)
	grpcports := make([]int, 0)
	for k, v := range lst {
		MyPrint(0, string(k), v)
		ports = append(ports, BasePort+k)
		grpcports = append(grpcports, GrpcBasePort+k)
	}
	return lst, ports, grpcports
}
