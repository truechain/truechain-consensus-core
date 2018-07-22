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
	"strconv"
	"time"

	"trueconsensus/common"

	pb "trueconsensus/fastchain/proto"
)

// GetDefaultGenesisBlock returns a default genesis block
func GetDefaultGenesisBlock() *pb.PbftBlock {
	txns := make([]*pb.Transaction, 0)
	genesisBlock := &pb.PbftBlock{}
	genesisBlockHeader := &pb.PbftBlockHeader{
		Number:     0,
		GasLimit:   5000,
		GasUsed:    0,
		Timestamp:  time.Now().Unix(),
		ParentHash: []byte(strconv.FormatInt(0, 16)),
		TxnsHash:   common.HashTxns(txns),
	}

	genesisBlock.Header = genesisBlockHeader

	return genesisBlock
}
