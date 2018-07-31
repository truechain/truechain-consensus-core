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
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/rpc"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"trueconsensus/common"

	pb "trueconsensus/fastchain/proto"

	"github.com/alecthomas/repr"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// const ServerPort = 40162
// const BufferSize = 4096 * 32
const (
	typePrePrepare = iota
	typePrepare    = iota
	typeCommit     = iota
	typeInit       = iota
	TypeRequest    = iota
	typeViewChange = iota
	typeNewView    = iota
	typeCheckpoint = iota
)

// TODO: change all the int to int64 in case of overflow

// ToGOB64 is a go binary encoder
func ToGOB64(m []byte) string {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(m)
	if err != nil {
		fmt.Println(`failed gob Encode`, err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

// FromGOB64 is a base64 decoder
func FromGOB64(str string) []byte {
	m := make([]byte, 0)
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println(`failed base64 Decode`, err)
	}
	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	err = d.Decode(&m)
	if err != nil {
		fmt.Println(`failed gob Decode`, err)
	}
	return m
}

// GetHash generates sha512 hash for a string
func GetHash(plaintext string) string {
	hasher := sha512.New()
	hasher.Write([]byte(plaintext))
	return hex.EncodeToString(hasher.Sum(nil))
}

// ApplyMsg keeps track of messages and their indices.
// It's used by both the server and client. Consider it as
// kind of like the common Z environment variable.
type ApplyMsg struct {
	Index   int
	Content interface{}
}

type clientMessageLogItem struct {
	clientID  int
	timestamp int64
}

type viewDictKey struct {
	seq int
	id  int
}

type nodeMsgLog struct {
	mu      sync.Mutex
	content map[int](map[int](map[int]Request))
}

func (nml *nodeMsgLog) get(typ int, seq int, id int) (Request, bool) {
	nml.mu.Lock()
	defer nml.mu.Unlock()
	_, ok := nml.content[typ]

	if !ok {
		nml.content[typ] = make(map[int](map[int]Request))

	}
	_, ok2 := nml.content[typ][seq]
	if !ok2 {
		nml.content[typ][seq] = make(map[int]Request)

	}
	val, ok3 := nml.content[typ][seq][id]
	if ok3 {
		return val, true
	}
	return Request{}, false
}

func (nml *nodeMsgLog) set(typ int, seq int, id int, req Request) {
	_, _ = nml.get(typ, seq, id) // in case there is an access error
	nml.mu.Lock()
	defer nml.mu.Unlock()
	nml.content[typ][seq][id] = req
}

type commitDictItem struct {
	numberOfCommit int
	committed      bool
	prepared       bool
	req            Request
}

type checkpointProofType Request

type viewDictItem struct {
	store   []Request
	holder1 int
	holder2 int
}

type viewDictType map[int]viewDictItem

type hellowSignature big.Int

type keyItem *ecdsa.PublicKey

// Node contains base properties of a Node
type Node struct {
	cfg      *Config
	mu       sync.Mutex
	clientMu sync.Mutex
	peers    []*rpc.Client
	port     int
	killFlag bool

	ListenReady chan bool
	SetupReady  chan bool

	EcdsaKey             *ecdsa.PrivateKey
	helloSignature       *hellowSignature
	connections          int
	ID                   int
	N                    int
	view                 int
	viewInUse            bool
	f                    int
	lowBound             int
	highBound            int
	Primary              int
	seq                  int
	lastExecuted         int
	lastStableCheckpoint int
	checkpointProof      []checkpointProofType
	checkpointInterval   int
	vmin                 int
	vmax                 int
	waiting              map[int]Request
	timeout              int
	clientBuffer         string
	active               map[DigType]ActiveItem
	prepared             map[int]Request
	prepDict             map[DigType]reqCounter
	commDict             map[DigType]reqCounter
	viewDict             viewDictType
	KeyDict              map[int]keyItem

	outputLog *os.File
	commitLog *os.File

	nodeMessageLog nodeMsgLog

	/// log
	clientMessageLog map[clientMessageLogItem]Request
	// nodeMessageLog	map[string](map[int](map[int]Request))

	committedBlock chan *pb.PbftBlock

	txPool  *TxPool
	genesis *pb.PbftBlock
	tc      *pb.TrueChain
}

type reqCounter struct {
	number   int
	prepared bool
	req      *Request
}

// Log - Final daily LOG that is signed by the BFT committee member
type Log struct {
	//TODO
}

// ActiveItem keeps track of currently active request and msg digest
type ActiveItem struct {
	req      *Request
	t        *time.Timer
	clientID int
}

// DigType is the message digest type
type DigType string

// MsgType contains the message of the request as part of RequestInner
type MsgType string

// RequestInner represents the core structure of a request
// Reqtype maps to following iota'ed constants
// ( typePrePrepare, typePrepare, typeCommit, typeInit,
//   typeRequest, typeViewChange, typeNewView, typeCheckpoint )
type RequestInner struct {
	ID        int
	Seq       int
	View      int
	Reqtype   int //or int?
	Msg       MsgType
	Block     *pb.PbftBlock
	Timestamp int64
	//outer *Request
}

// MsgSignature contains the EC coordinates
type MsgSignature struct {
	R *big.Int
	S *big.Int
}

// Request is the object juggled between the nodes
type Request struct {
	Inner RequestInner
	Dig   DigType
	Sig   MsgSignature
}

func (nd *Node) createRequest(reqType int, seq int, msg MsgType, block *pb.PbftBlock) Request {
	key := nd.EcdsaKey
	m := RequestInner{nd.ID, seq, nd.view, reqType, msg, block, -1} // , nil}
	req := Request{m, "", MsgSignature{nil, nil}}
	//req.Inner.outer = &req
	req.AddSig(key)
	return req
}

// AddSig signs message with private key of node
func (req *Request) AddSig(privKey *ecdsa.PrivateKey) {
	//common.MyPrint(1, "adding signature.\n")
	gob.Register(&RequestInner{})
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(req.Inner)
	if err != nil {
		common.MyPrint(3, "%s err:%s", `failed to encode!\n`, err.Error())
		return
	}

	s := GetHash(string(b.Bytes()))
	common.MyPrint(1, "digest %s.\n", string(s))
	req.Dig = DigType(s)
	if privKey != nil {
		sigr, sigs, err := ecdsa.Sign(rand.Reader, privKey, []byte(s))
		if err != nil {
			common.MyPrint(3, "%s", "Error signing.")
			return
		}

		req.Sig = MsgSignature{sigr, sigs}
	}
}

func (req *Request) verifyDig() bool {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(req.Inner)
	if err != nil {
		common.MyPrint(3, "%s err:%s", `failed to encode!`, err.Error())
		return false
	}

	s := GetHash(string(b.Bytes()))
	return s == string(req.Dig)
}

func (nd *Node) resetMsgDicts() {
	nd.nodeMessageLog = nodeMsgLog{}
	nd.prepDict = make(map[DigType]reqCounter)
	nd.commDict = make(map[DigType]reqCounter)
}

func (nd *Node) suicide() {
	nd.killFlag = true
}

// ProxyProcessPrePrepareArg holds context for committee, maps Request to client
type ProxyProcessPrePrepareArg struct {
	Req      Request
	ClientID int
}

// ProxyProcessPrePrepareReply is a stub atm, an ack of reply
type ProxyProcessPrePrepareReply struct {
}

// ProxyProcessPrepareArg holds client-request context for Prepare phase of PBFT
type ProxyProcessPrepareArg struct {
	Req      Request
	ClientID int
}

// ProxyProcessPrepareReply is a stub atm, an ack of reply to trigger Prepare
type ProxyProcessPrepareReply struct {
}

// ProxyProcessCommitArg holds Commit Phase declarative args
type ProxyProcessCommitArg struct {
	Req Request
}

// ProxyProcessCommitReply reply ack for PBFT commit phase
type ProxyProcessCommitReply struct {
}

// ProxyProcessViewChangeArg holds view change context from previous request
type ProxyProcessViewChangeArg struct {
	Req  Request
	From int
}

// ProxyProcessViewChangeReply holds view change reply context
type ProxyProcessViewChangeReply struct {
}

// ProxyProcessNewViewArg holds new view arg for client request
type ProxyProcessNewViewArg struct {
	Req      Request
	ClientID int
}

// ProxyProcessNewViewReply reply ack for new view request
type ProxyProcessNewViewReply struct {
}

// ProxyProcessCheckpointArg holds checkPoint args for PBFT checkpoints
type ProxyProcessCheckpointArg struct {
	Req      Request
	ClientID int
}

// ProxyProcessCheckpointReply ack for PBFT checkpoint
type ProxyProcessCheckpointReply struct {
}

func (nd *Node) broadcast(req Request) {
	//common.MyPrint(1, "[%d] Broadcast %v\n", nd.ID, req)
	switch req.Inner.Reqtype {
	// refine the following
	case typePrePrepare:
		arg := ProxyProcessPrePrepareArg{req, req.Inner.ID}
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessPrePrepareReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessPrePrepare", arg, &reply)
		break
	case typePrepare:
		arg := ProxyProcessPrepareArg{req, req.Inner.ID}
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessPrepareReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessPrepare", arg, &reply)
		break
	case typeCommit:
		arg := ProxyProcessCommitArg{req}
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessCommitReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessCommit", arg, &reply)
		break
	case typeViewChange:
		arg := ProxyProcessViewChangeArg{req, nd.ID} // a dummy from
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessViewChangeReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessViewChange", arg, &reply)
		break
	case typeNewView:
		arg := ProxyProcessNewViewArg{req, req.Inner.ID}
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessNewViewReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessNewView", arg, &reply)
		break
	case typeCheckpoint:
		arg := ProxyProcessCheckpointArg{req, req.Inner.ID}
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessCheckpointReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessCheckpoint", arg, &reply)
		break
	default:
		panic(1) // something bad

	}
}

// ProxyProcessPrePrepare trigger Pre-Prepare request
func (nd *Node) ProxyProcessPrePrepare(arg ProxyProcessPrePrepareArg, reply *ProxyProcessPrePrepareReply) error {
	common.MyPrint(2, "[%d] ProxyProcessPrePrepare %v\n", nd.ID, arg)
	nd.processPrePrepare(arg.Req, arg.ClientID) // we don't have return value here
	return nil
}

// ProxyProcessPrepare trigger Prepare phase of PBFT
func (nd *Node) ProxyProcessPrepare(arg ProxyProcessPrepareArg, reply *ProxyProcessPrepareReply) error {
	common.MyPrint(2, "[%d] ProxyProcessPrepare %v\n", nd.ID, arg.Req.Inner.Block.Header.Number)
	nd.processPrepare(arg.Req, arg.ClientID) // we don't have return value here
	return nil
}

// ProxyProcessCommit trigger Commit phase of PBFT
func (nd *Node) ProxyProcessCommit(arg ProxyProcessCommitArg, reply *ProxyProcessCommitReply) error {
	common.MyPrint(2, "[%d] ProxyProcessCommit %v\n", nd.ID, arg.Req.Inner.Block.Header.Number)
	nd.processCommit(arg.Req) // we don't have return value here
	return nil
}

// ProxyProcessViewChange trigger view change process request
func (nd *Node) ProxyProcessViewChange(arg ProxyProcessViewChangeArg, reply *ProxyProcessViewChangeReply) error {
	common.MyPrint(2, "[%d] ProxyProcessViewChange %v\n", nd.ID, arg)
	nd.processViewChange(arg.Req, arg.From)
	return nil
}

// ProxyProcessNewView triggers and frames a new view from current request
func (nd *Node) ProxyProcessNewView(arg ProxyProcessNewViewArg, reply *ProxyProcessNewViewReply) error {
	common.MyPrint(2, "[%d] ProxyProcessNewView %v\n", nd.ID, arg)
	nd.processNewView(arg.Req, arg.ClientID)
	return nil
}

// ProxyProcessCheckpoint processes checkpoint for PBFT ledger
func (nd *Node) ProxyProcessCheckpoint(arg ProxyProcessCheckpointArg, reply *ProxyProcessCheckpointReply) error {
	nd.processCheckpoint(arg.Req, arg.ClientID)
	return nil
}

// broadcast to all the peers
func (nd *Node) broadcastByRPC(rpcPath string, arg interface{}, reply *[]interface{}) {
	common.MyPrint(2, "[%d] Broadcasting to %s, %v\n", nd.ID, rpcPath, arg)
	divCallList := make([]*rpc.Call, 0)
	for ind := 0; ind < nd.N; ind++ {
		if ind == nd.ID {
			continue // skip the node itself
		}
		c := nd.peers[ind]
		divCallList = append(divCallList, c.Go(rpcPath, arg, (*reply)[ind], nil))
	}
	// synchronize
	for _, divCall := range divCallList {
		divCallDone := <-divCall.Done
		if divCallDone.Error != nil {
			common.MyPrint(3, "%s", "error happened in broadcasting "+rpcPath+"\n")
		}
	}
	//common.MyPrint(2, "[%d] Finished broadcasting to %s, %v\n", nd.ID, rpcPath, arg)
}

func (nd *Node) executeInOrder(req Request) {
	// Not sure if we need this inside the protocol. I tend to put this in application layer.
	waiting := true
	//r := req
	seq := req.Inner.Seq
	dig := DigType(req.Inner.Msg)
	nd.mu.Lock()
	ac := nd.active[dig]
	nd.mu.Unlock()
	if ac.req == nil {
		return
	}
	for seq == nd.lastExecuted+1 {
		waiting = false
		nd.lastExecuted++
		_, ok := nd.waiting[seq+1]
		if ok {
			seq++
			//r = rtemp
			nd.mu.Lock()
			delete(nd.waiting, seq)
			nd.mu.Unlock()
		}

	}

	if waiting {
		nd.mu.Lock()
		nd.waiting[req.Inner.Seq] = req
		nd.mu.Unlock()
	}
}

func (nd *Node) serverLoop() {
	common.MyPrint(1, "[%d] Entering server loop.\n", nd.ID)

	server := rpc.NewServer()
	// TODO: gives out the error:
	// rpc.Register: reply type of method "NewClientRequest" is not a pointer: "int"
	server.Register(nd)

	l, e := net.Listen("tcp", ":"+strconv.Itoa(nd.port))
	if e != nil {
		common.MyPrint(3, "listen error:%s", e.Error())
	}

	go server.Accept(l)

	nd.ListenReady <- true // trigger the connection
	common.MyPrint(1, "[%d] Ready to listen on %d.\n", nd.ID, nd.port)
}

func (nd *Node) clean() {
	// TODO: do clean work
}

func (nd *Node) processCheckpoint(req Request, clientID int) {
	// TODO: Add checkpoint support
	// Empty for now
}

func (nd *Node) isInClientLog(req Request) bool {
	_, ok := nd.clientMessageLog[clientMessageLogItem{req.Inner.ID, req.Inner.Timestamp}]
	return ok
}

func (nd *Node) addClientLog(req Request) {
	if !nd.isInClientLog(req) {
		nd.clientMessageLog[clientMessageLogItem{req.Inner.ID, req.Inner.Timestamp}] = req
	}
	return // We don't have logs for now
}

func (nd *Node) handleTimeout(dig DigType, view int) {
	nd.mu.Lock()
	if nd.view > view {
		nd.mu.Unlock()
		return
	}
	common.MyPrint(2, "%s", "Timeout triggered.\n")
	for k, v := range nd.active {
		if k != dig {
			v.t.Stop()
		}
	}
	nd.view++
	nd.viewInUse = false
	nd.mu.Unlock()

	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	e.Encode(len(nd.checkpointProof))
	for cp := range nd.checkpointProof {
		e.Encode(cp)
	}
	e.Encode(len(nd.prepared))
	for k := range nd.prepared {
		r, _ := nd.nodeMessageLog.get(typePrePrepare, k, nd.Primary) // old primary
		e.Encode(r)
		counter := 0
		for i := 1; i < nd.N; i++ {
			if counter == 2*nd.f {
				rtmp := Request{} // to make sure that alwyas N reqs are encoded
				rtmp.Inner = RequestInner{}
				rtmp.Inner.ID = -1
				e.Encode(rtmp)
				break
			}
			r, ok := nd.nodeMessageLog.get(typePrepare, k, i)
			if ok {
				e.Encode(r)
				counter++
				// TODO: add counter whenever we found a PREP
			}
		}
	}

	viewChange := nd.createRequest(typeViewChange, nd.lastStableCheckpoint, MsgType(b.Bytes()), nil)
	nd.broadcast(viewChange) // TODO: broadcast view change RPC path.
	nd.processViewChange(viewChange, 0)
}

// newClientRequest handles transaction request from client and broadcasts it to other PBFT nodes from primary replica
func (nd *Node) newClientRequest(req Request, clientID int) { // TODO: change to single arg and single reply
	if v, ok := nd.active[req.Dig]; ok {
		if v.req == nil {
			nd.mu.Lock()
			nd.active[req.Dig] = ActiveItem{&req, v.t, v.clientID}
			nd.mu.Unlock()
			if v2, ok2 := nd.commDict[req.Dig]; ok2 && v2.prepared {
				msg := v2.req
				nd.executeInOrder(*msg)
			}
			return
		}
	}

	nd.mu.Lock()
	if !nd.viewInUse {
		nd.mu.Unlock()
		return
	}
	nd.addClientLog(req)

	// Setup timer to track timeout on request
	reqTimer := time.NewTimer(time.Duration(nd.timeout) * time.Second)
	go func(t *time.Timer, r Request) {
		<-reqTimer.C
		nd.handleTimeout(req.Dig, req.Inner.View)
	}(reqTimer, req)

	// Add this request to map keeping track of currently active requests
	nd.active[req.Dig] = ActiveItem{&req, reqTimer, clientID}
	nd.mu.Unlock()

	// If node is primary replica broadcast request to other pbft nodes
	if nd.Primary == nd.ID {
		common.MyPrint(2, "[%d] Leader acked: %v.\n", nd.ID, req)
		nd.seq = nd.seq + 1
		m := nd.createRequest(typePrePrepare, nd.seq, MsgType(req.Dig), req.Inner.Block)
		nd.nodeMessageLog.set(m.Inner.Reqtype, m.Inner.Seq, m.Inner.ID, m)
		// write log
		nd.broadcast(m) // TODO: broadcast pre-prepare RPC path.
	}
}

func (nd *Node) initializeKeys() {
	gob.Register(&ecdsa.PrivateKey{})

	for i := 0; i < nd.N; i++ {
		pubkeyFile := fmt.Sprintf("sign%v.pub", i)
		fmt.Println("fetching file: ", pubkeyFile)
		pubKeyBytes, _ := common.FetchPublicKeyBytes(path.Join(nd.cfg.Logistics.KD, pubkeyFile))
		pubKey, _ := ethcrypto.UnmarshalPubkey(pubKeyBytes)
		nd.KeyDict[i] = pubKey
		if i == nd.ID {
			pemkeyFile := fmt.Sprintf("sign%v.pem", i)
			pemKey, _ := ethcrypto.LoadECDSA(path.Join(nd.cfg.Logistics.KD, pemkeyFile))
			nd.EcdsaKey = pemKey
			common.MyPrint(1, "Fetched private key")
		}
	}
}

func (nd *Node) incPrepDict(dig DigType) {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	if val, ok := nd.prepDict[dig]; ok {
		val.number++
		nd.prepDict[dig] = val
	} else {
		nd.prepDict[dig] = reqCounter{1, false, nil}
	}
}

func (nd *Node) incCommDict(dig DigType) {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	if val, ok := nd.commDict[dig]; ok {
		val.number++
		nd.commDict[dig] = val
	} else {
		nd.commDict[dig] = reqCounter{1, false, nil}
	}
}

func (nd *Node) checkPrepareMargin(dig DigType, seq int) bool {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	if val, ok := nd.prepDict[dig]; ok {
		if !val.prepared {
			fmt.Printf("prepare margin check number %d\n", val.number)
			if val.number >= 2*nd.f+1 {
				val, ok := nd.nodeMessageLog.get(typePrePrepare, seq, nd.Primary)
				common.MyPrint(2, "Dig: %v, val.Dig: %v\n", dig, DigType(val.Inner.Msg))
				if ok && DigType(val.Inner.Msg) == dig { // TODO: check the diff!
					if valt, okt := nd.prepDict[dig]; okt {
						valt.prepared = true
						nd.prepDict[dig] = valt
						return true
					}
				}
			}
		}
	}
	return false
}

func (nd *Node) checkCommittedMargin(dig DigType, req Request) bool {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	seq := req.Inner.Seq
	if val, ok := nd.commDict[dig]; ok {
		if !val.prepared {
			//fmt.Printf("commit margin check number %d\n", val.number)
			if val.number >= 2*nd.f+1 {
				val, ok := nd.nodeMessageLog.get(typePrePrepare, seq, nd.Primary)
				if ok && DigType(val.Inner.Msg) == dig { // TODO: check the diff!
					if valt, okt := nd.commDict[dig]; okt {
						valt.prepared = true
						valt.req = &req
						nd.commDict[dig] = valt
						return true
					}
				}
			}
		}
	}
	return false
}

// VerifySender verifies transaction sender by matching public key obtained from transaction signature with expected public key
// TODO This should match sender addresses later
func VerifySender(tx *pb.Transaction, n int) ([]byte, bool) {
	sig := tx.Data.Signature
	txPubkey, _ := ethcrypto.Ecrecover(tx.Data.Hash, sig)

	cfg := GetPbftConfig()
	pubKeyFile := fmt.Sprintf("sign%v.pub", n)
	fmt.Println("fetching file: ", pubKeyFile)
	pubKey, _ := common.FetchPublicKeyBytes(path.Join(cfg.Logistics.KD, pubKeyFile))

	if bytes.Equal(txPubkey, pubKey) {
		return pubKey, true
	}

	return nil, false
}

// VerifyBlockTxs verifies transactions in a block by checking sender and account nonce
func (nd *Node) VerifyBlockTxs(blk *pb.PbftBlock) bool {
	for _, tx := range blk.Txns {
		if _, ok := VerifySender(tx, nd.cfg.Network.N); !ok { // Verify if tx came from client
			return false
		}

		foundTx := make(chan *pb.Transaction, 1)
		txInPool := make(chan bool, 1)
		go func() {
			// Setup timer to track timeout on transaction to appear on pool
			timer := time.NewTimer(time.Duration(nd.timeout) * time.Second)
			go func(t *time.Timer) {
				<-timer.C
				txInPool <- false
			}(timer)

			txHash := common.BytesToHash(tx.Data.Hash)
			for {
				if t := nd.txPool.Get(txHash); t != nil {
					foundTx <- t
					txInPool <- true
					break
				}
			}
		}()

		okchan := make(chan bool, 1)
		go func() {
			found := <-txInPool
			if !found {
				okchan <- false // tx not found in txnpool after timeout
			}

			t := <-foundTx
			if t.Data.AccountNonce != tx.Data.AccountNonce {
				okchan <- false // tx account nonce from txPool didn't match that in block
			}

			okchan <- true
		}()

		if ok := <-okchan; !ok {
			return false
		}
	}

	return true
}

func (nd *Node) processPrePrepare(req Request, clientID int) {
	seq := req.Inner.Seq
	if val1, ok1 := nd.nodeMessageLog.content[typePrePrepare]; ok1 {
		if _, ok2 := val1[seq]; ok2 {
			return
		}
	}

	block := req.Inner.Block
	if !nd.VerifyBlockTxs(block) {
		common.MyPrint(3, "Couldn't verify transactions in block")
		return
	}

	if _, ok := nd.active[DigType(req.Inner.Msg)]; !ok {
		reqTimer := time.NewTimer(time.Duration(nd.timeout) * time.Second)

		go func(t *time.Timer, r Request) {
			<-reqTimer.C
			nd.handleTimeout(DigType(req.Inner.Msg), req.Inner.View)
		}(reqTimer, req)

		nd.mu.Lock()
		nd.active[DigType(req.Inner.Msg)] = ActiveItem{&req, reqTimer, clientID}
		nd.mu.Unlock()
	}

	nd.nodeMessageLog.set(req.Inner.Reqtype, req.Inner.Seq, req.Inner.ID, req)
	m := nd.createRequest(typePrepare, req.Inner.Seq, MsgType(req.Inner.Msg), req.Inner.Block) // TODO: check content!
	nd.nodeMessageLog.set(m.Inner.Reqtype, m.Inner.Seq, m.Inner.ID, m)
	nd.recordPBFT(m)
	nd.incPrepDict(DigType(req.Inner.Msg))
	nd.broadcast(m)
	if nd.checkPrepareMargin(DigType(req.Inner.Msg), req.Inner.Seq) { // TODO: check dig vs Inner.Msg
		nd.record("PREPARED sequence number " + strconv.Itoa(req.Inner.Seq) + "\n")
		m := nd.createRequest(typeCommit, req.Inner.Seq, req.Inner.Msg, req.Inner.Block) // TODO: check content
		nd.broadcast(m)
		nd.nodeMessageLog.set(m.Inner.Reqtype, m.Inner.Seq, m.Inner.ID, m)
		nd.incCommDict(DigType(req.Inner.Msg)) //TODO: check content
		nd.recordPBFT(m)
		nd.prepared[req.Inner.Seq] = m // or msg?
		if nd.checkCommittedMargin(DigType(req.Inner.Msg), m) {
			nd.record("COMMITED seq number " + strconv.Itoa(m.Inner.Seq) + "\n")
			nd.recordPBFTCommit(m)
			nd.executeInOrder(m)
		}
	}
}

func (nd *Node) recordCommit(content string) {
	if _, err := nd.commitLog.Write([]byte(content)); err != nil {
		panic(err)
	}
}

func (nd *Node) record(content string) {
	if _, err := nd.outputLog.Write([]byte(content)); err != nil {
		panic(err)
	}
}

func (nd *Node) recordPBFTCommit(req Request) {
	reqsummary := repr.String(&req)
	res := fmt.Sprintf("[%d] seq:%d from:%d view:%d %s\n", req.Inner.Reqtype, req.Inner.Seq, req.Inner.ID, req.Inner.View, reqsummary)
	nd.recordCommit(res)
}

func (nd *Node) recordPBFT(req Request) {
	reqsummary := repr.String(&req)
	res := fmt.Sprintf("[%d] seq:%d from:%d view:%d %s\n", req.Inner.Reqtype, req.Inner.Seq, req.Inner.ID, req.Inner.View, reqsummary)
	nd.record(res)
}

func (nd *Node) addNodeHistory(req Request) {
	nd.nodeMessageLog.set(req.Inner.Reqtype, req.Inner.Seq, req.Inner.ID, req)
}

func (nd *Node) processPrepare(req Request, clientID int) {
	nd.addNodeHistory(req)
	nd.incPrepDict(DigType(req.Inner.Msg))
	if nd.checkPrepareMargin(DigType(req.Inner.Msg), req.Inner.Seq) { // TODO: check dig vs Inner.Msg
		common.MyPrint(2, "[%d] Checked Margin\n", nd.ID)
		nd.record("PREPARED sequence number " + strconv.Itoa(req.Inner.Seq) + "\n")
		m := nd.createRequest(typeCommit, req.Inner.Seq, req.Inner.Msg, req.Inner.Block) // TODO: check content
		nd.broadcast(m)
		nd.nodeMessageLog.set(m.Inner.Reqtype, m.Inner.Seq, m.Inner.ID, m)
		nd.incCommDict(m.Dig) //TODO: check content
		nd.recordPBFT(m)
		nd.prepared[req.Inner.Seq] = m // or msg?
		if nd.checkCommittedMargin(m.Dig, m) {
			nd.record("COMMITED seq number " + strconv.Itoa(m.Inner.Seq) + "\n")
			nd.recordPBFTCommit(m)
			nd.executeInOrder(m)
		}
	}
}

// NewTrueChain creates a fresh blockchain
func (nd *Node) newTrueChain() {

	tc := &pb.TrueChain{}

	tc.Blocks = make([]*pb.PbftBlock, 0)
	tc.Blocks = append(tc.Blocks, nd.genesis)
	tc.LastBlockHeader = nd.genesis.Header

	nd.tc = tc
}

// AppendBlock appends a block to the blockchain
func (nd *Node) appendBlock(block *pb.PbftBlock) {
	nd.tc.Blocks = append(nd.tc.Blocks, block)
	nd.tc.LastBlockHeader = block.Header
}

func (nd *Node) processCommit(req Request) {
	nd.addNodeHistory(req)
	nd.incCommDict(DigType(req.Inner.Msg))
	if nd.checkCommittedMargin(DigType(req.Inner.Msg), req) {
		nd.committedBlock <- req.Inner.Block
		common.MyPrint(2, "[%d] Committed %v\n", nd.ID, req.Inner.Block.Header.Number)
		nd.record("COMMITTED seq number " + strconv.Itoa(req.Inner.Seq))
		nd.recordPBFTCommit(req)
		nd.executeInOrder(req)
	}
}

func (nd *Node) viewProcessCheckPoint(vchecklist *[]Request, lastCheckPoint int) bool {
	if lastCheckPoint == 0 {
		return true
	}
	if len(*vchecklist) <= 2*nd.f {
		return false
	}
	dig := (*vchecklist)[0].Dig
	for _, c := range *vchecklist {
		if c.Inner.Seq != lastCheckPoint || c.Dig != dig {
			return false
		}
	}
	return true
}

type viewDict map[int](map[int]Request) //map[viewDictKey]Request

func (nd *Node) verifyMsg(req Request) bool {
	key := nd.KeyDict[req.Inner.ID]
	// check the digest
	dv := req.verifyDig()
	// check the signature
	sc := ecdsa.Verify(key, []byte(req.Dig), req.Sig.R, req.Sig.S)
	return dv && sc
}

func (nd *Node) viewProcessPrepare(vPrepDict viewDict, vPreDict map[int]Request, lastCheckPoint int) (bool, int) {
	max := 0
	counter := make(map[int]int)
	for k1, v1 := range vPrepDict {
		if _, ok := vPreDict[k1]; !ok {
			return false, 0
		}
		reqTemp := vPreDict[k1]
		dig := reqTemp.Dig
		if !nd.verifyMsg(reqTemp) {
			return false, 0
		}
		for _, v2 := range v1 {
			if !nd.verifyMsg(v2) {
				return false, 0
			}
			if v2.Dig != dig {
				return false, 0
			}
			if reqTemp.Inner.ID == v2.Inner.ID {
				return false, 0 // cannot be sent from the same guy
			}
			if v2.Inner.Seq < lastCheckPoint {
				return false, 0
			}
			if v2.Inner.Seq > max {
				max = v2.Inner.Seq
			}
			if _, ok3 := counter[v2.Inner.Seq]; !ok3 {
				counter[v2.Inner.Seq] = 1
			} else {
				counter[v2.Inner.Seq]++
			}
		}
	}
	for k, v := range counter {
		if v < 2*nd.f {
			return false, 0
		}
		nd.addNodeHistory(vPreDict[k])
	}
	return true, max
}

func (nd *Node) processViewChange(req Request, from int) {
	common.MyPrint(1, "%s", "Receiveed a view change req from "+strconv.Itoa(req.Inner.ID))
	nd.addNodeHistory(req)
	newV := req.Inner.View
	if nd.view != req.Inner.View || newV < nd.view {
		return
	}
	/*
			# (NEVW, v+1, V, O) where V is set of valid VCHA, O is set of PRPR
		        #TODO (NO PIGGYBACK)
		        # determine the latest stable checkpoint
		        checkpoint = 0
		        # [seq][id] -> req
		        # for each view change message
		        #for r in self.view_dict[new_v]:

	*/
	//vcheckList := make([]Request, 0)
	m := req.Inner.Msg
	bufm := bytes.Buffer{}
	bufm.Write([]byte(m))
	//// extract msgs in m
	dec := gob.NewDecoder(&bufm)
	var lenCkPf, preparedLen int
	checkpointProofT := make([]Request, 0)
	vpreDict := make(map[int]Request)
	vprepDict := make(viewDict)
	dec.Decode(&lenCkPf)

	for i := 0; i < lenCkPf; i++ {
		ckpf := checkpointProofType{}
		dec.Decode(&ckpf)
		checkpointProofT = append(checkpointProofT, Request(ckpf))
	}
	dec.Decode(&preparedLen)
	for j := 0; j < preparedLen; j++ {
		r := Request{}
		dec.Decode(&r)
		for k := 0; k < nd.N; k++ {
			var rt Request
			dec.Decode(&rt)
			if rt.Inner.ID >= 0 { // so that this is not a dummy req
				switch rt.Inner.Reqtype {
				case typePrepare:
					if _, ok := vprepDict[rt.Inner.Seq]; !ok {
						vprepDict[rt.Inner.Seq] = make(map[int]Request)
					}
					vprepDict[rt.Inner.Seq][rt.Inner.ID] = rt
					break
				case typePrePrepare:
					vpreDict[rt.Inner.Seq] = rt
				}

			}
		}
	}
	rc1 := nd.viewProcessCheckPoint(&checkpointProofT, req.Inner.Seq) // fix the type of checkPointProofT
	rc2, maxm := nd.viewProcessPrepare(vprepDict, vpreDict, req.Inner.Seq)
	if rc1 && rc2 {
		if _, ok := nd.viewDict[newV]; !ok {
			reqList := make([]Request, 1)
			reqList[0] = req
			nd.viewDict[newV] = viewDictItem{reqList, 0, 0}
		} else {
			vTemp := nd.viewDict[newV]
			vTemp.store = append(vTemp.store, req)
			nd.viewDict[newV] = vTemp
		}
	}
	if nd.viewDict[newV].holder1 < req.Inner.Seq {
		tmp := nd.viewDict[newV]
		nd.viewDict[newV] = viewDictItem{tmp.store, req.Inner.Seq, tmp.holder2}
	}

	if nd.viewDict[newV].holder2 < maxm {
		tmp := nd.viewDict[newV]
		nd.viewDict[newV] = viewDictItem{tmp.store, tmp.holder1, maxm}
	}
	if (!nd.viewInUse || newV > nd.view) && len(nd.viewDict[newV].store) > 2*nd.f {
		// process and send the view req
		buf := bytes.Buffer{}
		e := gob.NewEncoder(&buf)
		reqList := make([]Request, 0)
		for i := range nd.viewDict[newV].store {
			//e.Encode(nd.viewDict[newV][0][i])
			reqList = append(reqList, nd.viewDict[newV].store[i])
		}
		for j := nd.viewDict[newV].holder1; j < nd.viewDict[newV].holder2; j++ {
			if j == 0 {
				continue
			}
			r, _ := nd.nodeMessageLog.get(typePrePrepare, j, nd.Primary)
			tmp := nd.createRequest(typePrePrepare, j, r.Inner.Msg, r.Inner.Block)
			reqList = append(reqList, tmp)
			//e.Encode(tmp)
		}
		e.Encode(reqList)
		out := nd.createRequest(typeNewView, 0, MsgType(buf.Bytes()), nil)
		nd.viewInUse = true
		nd.Primary = nd.view % nd.N
		nd.active = make(map[DigType]ActiveItem)
		nd.resetMsgDicts()
		nd.clientMessageLog = make(map[clientMessageLogItem]Request)
		nd.prepared = make(map[int]Request)
		nd.seq = nd.viewDict[newV].holder2
		nd.broadcast(out)
	}
}

func (nd *Node) newViewProcessPrePrepare(prprList []Request) bool {
	for _, r := range prprList {
		if nd.verifyMsg(r) && r.verifyDig() {
			out := nd.createRequest(typePrepare, r.Inner.Seq, r.Inner.Msg, r.Inner.Block)
			nd.broadcast(out)
		} else {
			return false
		}
	}
	return true
}

// newViewProcessView returns whether the vchangeList is genuine
// based on checks off of digest and message verifications
func (nd *Node) newViewProcessView(vchangeList []Request) bool {
	for _, r := range vchangeList {
		if !(nd.verifyMsg(r) && r.verifyDig()) {
			return false
		}
	}
	return true
}

// processNewView receives the request and client ID in order to
// handle consisteny across view change requests and triggers
// new PrePrepare phases accordingly
func (nd *Node) processNewView(req Request, clientID int) {
	m := req.Inner.Msg
	bufcurrent := bytes.Buffer{}
	bufcurrent.Write([]byte(m))
	e := gob.NewDecoder(&bufcurrent)
	vchangeList := make([]Request, 0)
	prprList := make([]Request, 0)
	//counter := 0
	reqList := make([]Request, 0)
	e.Decode(&reqList)
	for _, r := range reqList {
		if r.verifyDig() && nd.verifyMsg(r) {
			switch r.Inner.Reqtype {
			case typeViewChange:
				vchangeList = append(vchangeList, r)
				break
			case typePrePrepare:
				prprList = append(prprList, r)
				break
			}
		}
	}
	if !nd.newViewProcessView(vchangeList) {
		common.MyPrint(3, "%s", "Failed view change")
		return
	}
	if req.Inner.View >= nd.view {
		nd.view = req.Inner.View
		nd.viewInUse = true
		nd.Primary = nd.view % nd.N
		nd.active = make(map[DigType]ActiveItem)
		nd.resetMsgDicts()
		nd.clientMessageLog = make(map[clientMessageLogItem]Request)
		nd.prepared = make(map[int]Request)
		nd.newViewProcessPrePrepare(prprList)
		common.MyPrint(2, "%s", "New View accepted")
	}
}

// isInNodeLog ??
func (nd *Node) isInNodeLog() bool {
	// deprecated
	return false
}

// processRequest ??
func (nd *Node) processRequest() {
	// we use RPC so we don't need a central msg receiver
}

// beforeShutdown closes output log file buffer
func (nd *Node) beforeShutdown() {
	nd.outputLog.Close()
}

// setupConnections dials RPC connections to nodes from the client
func (nd *Node) setupConnections() {
	<-nd.SetupReady
	common.MyPrint(1, "[%d] Begin to setup RPC connections.\n", nd.ID)
	peers := make([]*rpc.Client, nd.cfg.Network.N)
	for i := 0; i < nd.cfg.Network.N; i++ {
		cl, err := rpc.Dial("tcp", nd.cfg.Network.IPList[i]+":"+strconv.Itoa(nd.cfg.Network.Ports[i]))
		if err != nil {
			common.MyPrint(3, "RPC error.\n")
		}
		peers[i] = cl
	}
	nd.peers = peers
	common.MyPrint(1, "[%d] Setup finished\n", nd.ID)
}

// createInternalPbftReq wraps a transaction request from client for internal rpc communication between pbft nodes
func (nd *Node) createInternalPbftReq(proposedBlock *pb.PbftBlock) Request {
	req := Request{}
	reqInner := RequestInner{}

	reqInner.ID = nd.cfg.N // client-id
	reqInner.Seq = 0
	reqInner.View = 0
	reqInner.Reqtype = TypeRequest // client request
	reqInner.Block = proposedBlock
	reqInner.Timestamp = time.Now().Unix()

	req.Inner = reqInner

	req.AddSig(nd.EcdsaKey)

	return req
}

func (nd *Node) createBlockAndBroadcast() {
	// This go routine creates the slices of block-size length and creates the block
	// for broadcasting once the previous block has been committed
	blockSize := nd.cfg.Blocksize
	go func() {
		for {
			if nd.txPool.GetTxCount() < blockSize {
				continue
			}

			blockTxs := make([]*pb.Transaction, 0)
			var gasUsed int64
			gasUsed = 0
			for i := 0; i < blockSize; i++ {
				tx := nd.txPool.priced.Get()
				blockTxs = append(blockTxs, tx)
				gasUsed = gasUsed + tx.Data.Price
				nd.txPool.Remove(common.BytesToHash(tx.Data.Hash))
				common.MyPrint(4, "Transacion count is %d", nd.txPool.GetTxCount())
			}
			parentHash := common.HashBlockHeader(nd.tc.LastBlockHeader)
			txsHash := common.HashTxns(blockTxs)
			header := NewPbftBlockHeader(nd.tc.LastBlockHeader.Number+1, 5000, int64(gasUsed), parentHash, txsHash)

			block := NewPbftBlock(header, blockTxs)

			req := nd.createInternalPbftReq(block)
			nd.newClientRequest(req, nd.cfg.N)
		}
	}()
}

// Make registers all node config objects,
func Make(cfg *Config, me int, port int, view int) *Node {
	gob.Register(&ApplyMsg{})
	gob.Register(&RequestInner{})
	gob.Register(&Request{})
	gob.Register(&checkpointProofType{})
	gob.Register(&ProxyProcessPrePrepareArg{})
	gob.Register(&ProxyProcessPrePrepareReply{})
	gob.Register(&ProxyProcessCheckpointArg{})
	gob.Register(&ProxyProcessCheckpointReply{})
	gob.Register(&ProxyProcessCommitArg{})
	gob.Register(&ProxyProcessCommitReply{})
	gob.Register(&ProxyProcessNewViewArg{})
	gob.Register(&ProxyProcessNewViewReply{})
	gob.Register(&ProxyProcessPrepareArg{})
	gob.Register(&ProxyProcessPrepareReply{})
	gob.Register(&ProxyProcessViewChangeArg{})
	gob.Register(&ProxyProcessViewChangeReply{})

	nd := &Node{}
	//rpc.Register(nd)
	nd.cfg = cfg
	nd.N = cfg.Network.N
	nd.f = (nd.N - 1) / 3
	common.MyPrint(2, "Going to tolerate %d adversaries\n", nd.f)
	nd.lowBound = 0
	nd.highBound = 0
	nd.view = view
	nd.viewInUse = true
	nd.Primary = view % nd.cfg.Network.N
	nd.port = port
	nd.seq = 0
	nd.lastExecuted = 0
	nd.lastStableCheckpoint = 0
	nd.checkpointProof = make([]checkpointProofType, 0)
	nd.timeout = 600
	//nd.mu = sync.Mutex{}
	nd.clientBuffer = ""
	//nd.clientMu = sync.Mutex{}
	nd.KeyDict = make(map[int]keyItem)
	nd.checkpointInterval = 100

	nd.vmin = 0
	nd.vmax = 0
	nd.waiting = nil
	nd.killFlag = false
	nd.ID = me
	nd.active = make(map[DigType]ActiveItem)
	nd.prepDict = make(map[DigType]reqCounter)
	nd.commDict = make(map[DigType]reqCounter)
	nd.nodeMessageLog = nodeMsgLog{}
	nd.nodeMessageLog.content = make(map[int](map[int](map[int]Request)))
	nd.prepared = make(map[int]Request)
	common.MyPrint(2, "Initial Node Config %+v\n", nd)
	nd.cfg.Logistics.LD = path.Join(GetCWD(), "logs/")
	// kfpath := path.Join(cfg.Logistics.KD, filename)

	nd.genesis = GetDefaultGenesisBlock()
	common.MyPrint(0, "Genesis block generated: %x\n\n", nd.genesis.Header.TxnsHash)

	nd.newTrueChain()

	// Buffered channel of size 1 contains the newly committed block to be added to
	// the blockchain by pbft-server
	nd.committedBlock = make(chan *pb.PbftBlock, 1)

	// Create empty transaction pool
	nd.txPool = newTxPool()

	common.MakeDirIfNot(nd.cfg.Logistics.LD) //handles 'already exists'
	fi, err := os.Create(path.Join(nd.cfg.Logistics.LD, "PBFTLog"+strconv.Itoa(nd.ID)+".txt"))
	if err == nil {
		nd.outputLog = fi
	} else {
		panic(err)
	}
	fi2, err2 := os.Create(path.Join(nd.cfg.Logistics.LD, "PBFTBuffer"+strconv.Itoa(nd.ID)+".txt"))
	if err2 == nil {
		nd.commitLog = fi2
	} else {
		panic(err)
	}

	nd.clientMessageLog = make(map[clientMessageLogItem]Request)
	nd.initializeKeys()
	nd.ListenReady = make(chan bool)
	nd.SetupReady = make(chan bool)
	nd.waiting = make(map[int]Request)
	go nd.setupConnections()

	go nd.serverLoop()

	if nd.ID == nd.Primary {
		nd.createBlockAndBroadcast()
	}

	go func() {
		for {
			block := <-nd.committedBlock
			nd.appendBlock(block)
		}
	}()

	return nd
}
