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
	"container/heap"
	"sync"

	pb "trueconsensus/fastchain/proto"
)

// nonceHeap is a heap.Interface implementation over 64bit unsigned integers for
// retrieving sorted transactions from the possibly gapped future queue.
type nonceHeap []uint64

func (h nonceHeap) Len() int           { return len(h) }
func (h nonceHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h nonceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nonceHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *nonceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// iterating over the contents in a nonce-incrementing way.
type txSortedMap struct {
	items map[uint64]*pb.Transaction // Hash map storing the transaction data
	index *nonceHeap                 // Heap of nonces of all the stored transactions (non-strict mode)
}

// newTxSortedMap creates a new nonce-sorted transaction map.
func newTxSortedMap() *txSortedMap {
	return &txSortedMap{
		items: make(map[uint64]*pb.Transaction),
		index: new(nonceHeap),
	}
}

// Get retrieves the current transactions associated with the given nonce.
func (m *txSortedMap) Get(nonce uint64) *pb.Transaction {
	return m.items[nonce]
}

// Put inserts a new transaction into the map, also updating the map's nonce
// index. If a transaction already exists with the same nonce, it's overwritten.
func (m *txSortedMap) Put(txn *pb.Transaction) {
	nonce := txn.Data.GetAccountNonce()
	if m.items[nonce] == nil {
		heap.Push(m.index, nonce)
	}
	m.items[nonce] = txn
}

// priceHeap is a heap.Interface implementation over transactions for retrieving
// price-sorted transactions to discard when the pool fills up.
type priceHeap []*pb.Transaction

func (h priceHeap) Len() int      { return len(h) }
func (h priceHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h priceHeap) Less(i, j int) bool {
	return h[i].Data.GetPrice() < h[j].Data.GetPrice()
}

func (h *priceHeap) Push(x interface{}) {
	*h = append(*h, x.(*pb.Transaction))
}

func (h *priceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// txPricedList is a price-sorted heap to allow operating on transactions pool
// contents in a price-incrementing way.
type txPricedList struct {
	all   *txLookup  // Pointer to the map of all transactions
	items *priceHeap // Heap of prices of all the stored transactions
	lock  sync.RWMutex
}

// newTxPricedList creates a new price-sorted transaction heap.
func newTxPricedList(all *txLookup) *txPricedList {
	return &txPricedList{
		all:   all,
		items: new(priceHeap),
	}
}

// Put inserts a new transaction into the heap.
func (l *txPricedList) Put(tx *pb.Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	heap.Push(l.items, tx)
}

// Get returns a new transaction based on the price associated
func (l *txPricedList) Get() *pb.Transaction {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return heap.Pop(l.items).(*pb.Transaction)
}

// Len returns number of items in priced list heap
func (l *txPricedList) Len() int {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.items.Len()
}
