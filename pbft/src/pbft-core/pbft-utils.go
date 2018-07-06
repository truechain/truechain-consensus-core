package pbft

import (
	pb "pbft-core/fastchain"

	"github.com/ethereum/go-ethereum/trie"
	"github.com/golang/protobuf/proto"
)

// HashTxns returns a hash of all the transactions using Merkle Patricia tries
func HashTxns(txns []pb.Transaction) []byte {
	trie := new(trie.Trie)
	for i, txn := range txns {
		val, _ := proto.Marshal(&txn)
		trie.Update(proto.EncodeVarint(uint64(i)), val)
	}

	return trie.Hash().Bytes()
}
