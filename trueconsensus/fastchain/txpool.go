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
	"sync"

	"trueconsensus/common"
	pb "trueconsensus/fastchain/proto"
)

// TxPool defines the transaction pool
type TxPool struct {
	txMap  map[common.Hash]*txSortedMap // Map of account to transactions sorted according to account nonce
	all    *txLookup                    // All transactions to allow lookups
	priced *txPricedList                // All transactions sorted by price
	count  int                          // Maintains a count of the number of transactions
	lock   sync.RWMutex
}

func newTxPool() *TxPool {
	txPool := &TxPool{}
	txPool.txMap = make(map[common.Hash]*txSortedMap)
	txPool.all = newTxLookup()
	txPool.priced = newTxPricedList(txPool.all)
	txPool.count = 0

	return txPool
}

// Add adds a new transaction from sender account to the transaction pool
func (tp *TxPool) Add(tx *pb.Transaction, sender []byte) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	senderHash := common.BytesToHash(sender)
	// If new sender add new sender entry to txmap
	if txs, _ := tp.txMap[senderHash]; txs == nil {
		txs = newTxSortedMap()
		tp.txMap[senderHash] = txs
	}

	tp.txMap[senderHash].Put(tx)

	tp.all.Add(tx)
	tp.priced.Put(tx)

	tp.count++
}

// Get returns a transaction if it is contained in the pool
// and nil otherwise.
func (tp *TxPool) Get(txHash common.Hash) *pb.Transaction {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	return tp.all.Get(txHash)
}

// Remove removes a transaction from the transaction pool
func (tp *TxPool) Remove(txHash common.Hash) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	tp.all.Remove(txHash)

	tp.count--
}

// GetTxCount returns the number of transactions currently in the transaction pool
func (tp *TxPool) GetTxCount() int {
	tp.lock.RLock()
	defer tp.lock.RUnlock()

	return tp.count
}

// txLookup defines a sructure that enables lookup for transactions
type txLookup struct {
	all  map[common.Hash]*pb.Transaction
	lock sync.RWMutex
}

// newTxLookup returns a new txLookup structure.
func newTxLookup() *txLookup {
	return &txLookup{
		all: make(map[common.Hash]*pb.Transaction),
	}
}

// Range calls f on each key and value present in the map.
func (t *txLookup) Range(f func(hash common.Hash, tx *pb.Transaction) bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	for key, value := range t.all {
		if !f(key, value) {
			break
		}
	}
}

// Get returns a transaction if it exists in the lookup, or nil if not found.
func (t *txLookup) Get(hash common.Hash) *pb.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.all[hash]
}

// Count returns the current number of items in the lookup.
func (t *txLookup) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.all)
}

// Add adds a transaction to the lookup.
func (t *txLookup) Add(tx *pb.Transaction) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.all[common.BytesToHash(tx.Data.Hash)] = tx
}

// Remove removes a transaction from the lookup.
func (t *txLookup) Remove(hash common.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()

	delete(t.all, hash)
}
