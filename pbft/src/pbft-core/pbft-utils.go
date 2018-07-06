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
	pb "pbft-core/fastchain"

	"github.com/ethereum/go-ethereum/trie"
	"github.com/golang/protobuf/proto"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// HashTxns returns a hash of all the transactions using Merkle Patricia tries
func HashTxns(txns []*pb.Transaction) []byte {
	trie := new(trie.Trie)
	for i, txn := range txns {
		val, _ := proto.Marshal(txn)
		trie.Update(proto.EncodeVarint(uint64(i)), val)
	}

	return trie.Hash().Bytes()
}

// HashHeader returns a hash for a block header
func HashBlockHeader(header *pb.PbftBlockHeader) []byte {
	headerData, _ := proto.Marshal(header)
	return ethcrypto.Keccak256(headerData)
}
