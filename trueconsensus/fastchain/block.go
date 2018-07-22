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
	"time"

	pb "trueconsensus/fastchain/proto"
)

// NewPbftBlockHeader creates and returns a new block header
func NewPbftBlockHeader(n, gasLimit, gasUsed int64, parentHash, txnsHash []byte) *pb.PbftBlockHeader {
	header := &pb.PbftBlockHeader{
		Number:     n,
		GasLimit:   gasLimit,
		GasUsed:    gasUsed,
		Timestamp:  time.Now().Unix(),
		ParentHash: parentHash,
		TxnsHash:   txnsHash,
	}

	return header
}

// NewPbftBlock creates and returns a new block
func NewPbftBlock(header *pb.PbftBlockHeader, txns []*pb.Transaction) *pb.PbftBlock {
	block := &pb.PbftBlock{}
	block.Header = header
	block.Txns = txns

	return block
}
