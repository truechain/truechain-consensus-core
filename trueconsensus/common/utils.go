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

package common

import (
	"fmt"
	"math/big"
	"math/rand"
	"reflect"

	pb "trueconsensus/fastchain/proto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/golang/protobuf/proto"
	// "trueconsensus/fastchain"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"crypto/ecdsa"
	"io/ioutil"

	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"github.com/fatih/color"
	"io"
	"os"
	// "strings"
)

// HashLength defines length of hashes (32 bytes)
const HashLength = 32

var hashT = reflect.TypeOf(Hash{})

// Hash defines a hash of 32 bytes
type Hash [32]byte

// BytesToHash sets b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// BigToHash sets byte representation of b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }

// HexToHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToHash(s string) Hash { return BytesToHash(common.FromHex(s)) }

// Bytes gets the byte representation of the underlying hash.
func (h Hash) Bytes() []byte { return h[:] }

// Big converts a hash to a big integer.
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }

// Hex converts a hash to a hex string.
func (h Hash) Hex() string { return hexutil.Encode(h[:]) }

// TerminalString implements log.TerminalStringer, formatting a string for console
// output during logging.
func (h Hash) TerminalString() string {
	return fmt.Sprintf("%x…%x", h[:3], h[29:])
}

// String implements the stringer interface and is used also by the logger when
// doing full logging into a file.
func (h Hash) String() string {
	return h.Hex()
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (h Hash) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), h[:])
}

// UnmarshalText parses a hash in hex syntax.
func (h *Hash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Hash", input, h[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (h *Hash) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(hashT, input, h[:])
}

// MarshalText returns the hex representation of h.
func (h Hash) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// SetBytes sets the hash to the value of b.
// If b is larger than len(h), b will be cropped from the left.
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// Generate implements testing/quick.Generator.
func (h Hash) Generate(rand *rand.Rand, size int) reflect.Value {
	m := rand.Intn(len(h))
	for i := len(h) - 1; i > m; i-- {
		h[i] = byte(rand.Uint32())
	}
	return reflect.ValueOf(h)
}

// HashTxns returns a hash of all the transactions using Merkle Patricia tries
func HashTxns(txns []*pb.Transaction) []byte {
	trie := new(trie.Trie)
	for i, txn := range txns {
		val, _ := proto.Marshal(txn)
		trie.Update(proto.EncodeVarint(uint64(i)), val)
	}

	return trie.Hash().Bytes()
}

// HashBlockHeader returns a hash for a block header
func HashBlockHeader(header *pb.PbftBlockHeader) []byte {
	headerData, _ := proto.Marshal(header)
	return ethcrypto.Keccak256(headerData)
}

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

	// if t >= OutputThreshold {
	// }
	switch t {
	case 0: // info
		fmt.Printf("[ ]"+format+"\n", args...)
		break
	case 1: // emphasized
		fmt.Printf(blue("[.]")+format+"\n", args...)
		break
	case 2: // warning
		fmt.Printf(yellow("[!]")+format+"\n", args...)
		break
	case 3: // error
		fmt.Printf(red("[x]")+format+"\n", args...)
		break
	case 4: // magical shit
		fmt.Printf(deepPurple("[?]")+format+"\n", args...)
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

// FetchPublicKeyBytes fetches ECDSA public key in []byte form
func FetchPublicKeyBytes(kpath string) ([]byte, error) {
	buf := make([]byte, 130) //  encoded key length is 1 + 2 * byteLen(publicKey). Also buf len should be even
	fd, err := os.Open(kpath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	if _, err := io.ReadFull(fd, buf); err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}

	return key, nil
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
