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

import "time"
import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/alecthomas/repr"
	"math/big"
	"net"
	"net/rpc"
	"os"
	"path"
	"strconv"
	"sync"
)

// const ServerPort = 40162
// const BufferSize = 4096 * 32
const (
	typePrePrepare = iota
	typePrepare    = iota
	typeCommit     = iota
	typeInit       = iota
	typeRequest    = iota
	typeViewChange = iota
	typeNewView    = iota
	typeCheckpoint = iota
)

// TODO: change all the int to int64 in case of overflow

// go binary encoder
func ToGOB64(m []byte) string {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(m)
	if err != nil {
		fmt.Println(`failed gob Encode`, err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

// go binary decoder
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

func getHash(plaintext string) string {
	hasher := sha512.New()
	hasher.Write([]byte(plaintext))
	return hex.EncodeToString(hasher.Sum(nil))
}

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

type DigType string

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

type Node struct {
	cfg          Config
	mu           sync.Mutex
	clientMu     sync.Mutex
	peers        []*rpc.Client
	port         int
	max_requests int
	kill_flag    bool

	ListenReady chan bool
	SetupReady  chan bool

	ecdsaKey             *ecdsa.PrivateKey
	helloSignature       *hellowSignature
	connections          int
	id                   int
	N                    int
	view                 int
	viewInUse            bool
	f                    int
	lowBound             int
	highBound            int
	primary              int
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
	keyDict              map[int]keyItem

	outputLog *os.File
	commitLog *os.File

	nodeMessageLog nodeMsgLog

	/// log
	clientMessageLog map[clientMessageLogItem]Request
	// nodeMessageLog	map[string](map[int](map[int]Request))

	/// apply channel

	applyCh chan ApplyMsg // Make sure that the client keeps getting the content out of applyCh
}

type reqCounter struct {
	number   int
	prepared bool
	req      *Request
}

type Log struct {
	//TODO
}

type ActiveItem struct {
	req      *Request
	t        *time.Timer
	clientID int
}

type MsgType string

type RequestInner struct {
	Id        int
	Seq       int
	View      int
	Reqtype   int //or int?
	Msg       MsgType
	Timestamp int64
	//outer *Request
}

type MsgSignature struct {
	R *big.Int
	S *big.Int
}

type Request struct {
	Inner RequestInner
	Dig   DigType
	Sig   MsgSignature
}

func (nd *Node) createRequest(reqType int, seq int, msg MsgType) Request {
	key := nd.ecdsaKey
	m := RequestInner{nd.id, seq, nd.view, reqType, msg, -1} // , nil}
	req := Request{m, "", MsgSignature{nil, nil}}
	//req.Inner.outer = &req
	req.addSig(key)
	return req
}

func (req *Request) addSig(privKey *ecdsa.PrivateKey) {
	//MyPrint(1, "adding signature.\n")
	gob.Register(&RequestInner{})
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(req.Inner)
	if err != nil {
		MyPrint(3, "%s", `failed to encode!\n`)
		fmt.Println(err)
	}
	s := getHash(string(b.Bytes()))
	MyPrint(1, "digest %s.\n", string(s))
	req.Dig = DigType(s)
	if privKey != nil {
		sigr, sigs, err := ecdsa.Sign(rand.Reader, privKey, []byte(s))
		if err != nil {
			MyPrint(3, "%s", "Error signing.")
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
		MyPrint(3, "%s", `failed to encode!`)
	}
	s := getHash(string(b.Bytes()))
	return s == string(req.Dig)
}

func (nd *Node) resetMsgDicts() {
	nd.nodeMessageLog = nodeMsgLog{}
	nd.prepDict = make(map[DigType]reqCounter)
	nd.commDict = make(map[DigType]reqCounter)
}

func (nd *Node) execute(am ApplyMsg) {
	// TODO: add msg to applyCh, should be executed in a separate go routine
	// TODO: add we probably want to keep a log for this
	nd.applyCh <- am
}

func (nd *Node) suicide() {
	nd.kill_flag = true
}

type ProxyProcessPrePrepareArg struct {
	Req      Request
	ClientID int
}

type ProxyProcessPrePrepareReply struct {
}

type ProxyNewClientRequestArg struct {
	Req      Request
	ClientID int
}

type ProxyNewClientRequestReply struct {
}

type ProxyProcessPrepareArg struct {
	Req      Request
	ClientID int
}

type ProxyProcessPrepareReply struct {
}

type ProxyProcessCommitArg struct {
	Req Request
}

type ProxyProcessCommitReply struct {
}

type ProxyProcessViewChangeArg struct {
	Req  Request
	From int
}

type ProxyProcessViewChangeReply struct {
}

type ProxyProcessNewViewArg struct {
	Req      Request
	ClientID int
}

type ProxyProcessNewViewReply struct {
}

type ProxyProcessCheckpointArg struct {
	Req      Request
	ClientID int
}

type ProxyProcessCheckpointReply struct {
}

func (nd *Node) broadcast(req Request) {
	//MyPrint(1, "[%d] Broadcast %v\n", nd.id, req)
	switch req.Inner.Reqtype {
	// refine the following
	case typePrePrepare:
		arg := ProxyProcessPrePrepareArg{req, req.Inner.Id}
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessPrePrepareReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessPrePrepare", arg, &reply)
		break
	case typeRequest:
		arg := ProxyNewClientRequestArg{req, req.Inner.Id}
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyNewClientRequestReply{}
		}
		nd.broadcastByRPC("Node.ProxyNewClientRequest", arg, &reply)
		break
	case typePrepare:
		arg := ProxyProcessPrepareArg{req, req.Inner.Id}
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
		arg := ProxyProcessViewChangeArg{req, nd.id} // a dummy from
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessViewChangeReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessViewChange", arg, &reply)
		break
	case typeNewView:
		arg := ProxyProcessNewViewArg{req, req.Inner.Id}
		reply := make([]interface{}, nd.N)
		for k := 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessNewViewReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessNewView", arg, &reply)
		break
	case typeCheckpoint:
		arg := ProxyProcessCheckpointArg{req, req.Inner.Id}
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

func (nd *Node) ProxyProcessPrePrepare(arg ProxyProcessPrePrepareArg, reply *ProxyProcessPrePrepareReply) error {
	MyPrint(2, "[%d] ProxyProcessPrePrepare %v\n", nd.id, arg)
	nd.processPrePrepare(arg.Req, arg.ClientID) // we don't have return value here
	return nil
}

func (nd *Node) ProxyNewClientRequest(arg ProxyNewClientRequestArg, reply *ProxyNewClientRequestReply) error {
	MyPrint(2, "New Client Request called.\n")
	nd.newClientRequest(arg.Req, arg.ClientID) // we don't have return value here
	return nil
}

func (nd *Node) ProxyProcessPrepare(arg ProxyProcessPrepareArg, reply *ProxyProcessPrepareReply) error {
	MyPrint(2, "[%d] ProxyProcessPrepare %v\n", nd.id, arg)
	nd.processPrepare(arg.Req, arg.ClientID) // we don't have return value here
	return nil
}

func (nd *Node) ProxyProcessCommit(arg ProxyProcessCommitArg, reply *ProxyProcessCommitReply) error {
	MyPrint(2, "[%d] ProxyProcessCommit %v\n", nd.id, arg)
	nd.processCommit(arg.Req) // we don't have return value here
	return nil
}

func (nd *Node) ProxyProcessViewChange(arg ProxyProcessViewChangeArg, reply *ProxyProcessViewChangeReply) error {
	MyPrint(2, "[%d] ProxyProcessViewChange %v\n", nd.id, arg)
	nd.processViewChange(arg.Req, arg.From)
	return nil
}

func (nd *Node) ProxyProcessNewView(arg ProxyProcessNewViewArg, reply *ProxyProcessNewViewReply) error {
	MyPrint(2, "[%d] ProxyProcessNewView %v\n", nd.id, arg)
	nd.processNewView(arg.Req, arg.ClientID)
	return nil
}

func (nd *Node) ProxyProcessCheckpoint(arg ProxyProcessCheckpointArg, reply *ProxyProcessCheckpointReply) error {
	nd.processCheckpoint(arg.Req, arg.ClientID)
	return nil
}

// broadcast to all the peers
func (nd *Node) broadcastByRPC(rpcPath string, arg interface{}, reply *[]interface{}) {
	MyPrint(2, "[%d] Broadcasting to %s, %v\n", nd.id, rpcPath, arg)
	divCallList := make([]*rpc.Call, 0)
	for ind := 0; ind < nd.N; ind++ {
		if ind == nd.id {
			continue // skip the node itself
		}
		c := nd.peers[ind]
		divCallList = append(divCallList, c.Go(rpcPath, arg, (*reply)[ind], nil))
	}
	// synchronize
	for _, divCall := range divCallList {
		divCallDone := <-divCall.Done
		if divCallDone.Error != nil {
			MyPrint(3, "%s", "error happened in broadcasting "+rpcPath+"\n")
		}
	}
	//MyPrint(2, "[%d] Finished broadcasting to %s, %v\n", nd.id, rpcPath, arg)
}

func (nd *Node) executeInOrder(req Request) {
	// Not sure if we need this inside the protocol. I tend to put this in application layer.
	waiting := true
	r := req
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
		nd.execute(ApplyMsg{seq, r})
		nd.lastExecuted++
		rtemp, ok := nd.waiting[seq+1]
		if ok {
			seq++
			r = rtemp
			delete(nd.waiting, seq)
		}

	}

	if waiting {
		nd.waiting[req.Inner.Seq] = req
	}
}

func (nd *Node) serverLoop() {
	MyPrint(1, "[%d] Entering server loop.\n", nd.id)

	server := rpc.NewServer()
	server.Register(nd)

	l, e := net.Listen("tcp", ":"+strconv.Itoa(nd.port))
	if e != nil {
		MyPrint(3, "listen error:", e)
	}

	go server.Accept(l)

	nd.ListenReady <- true // trigger the connection
	MyPrint(1, "[%d] Ready to listen on %d.\n", nd.id, nd.port)
}

func (nd *Node) clean() {
	// TODO: do clean work
}

func (nd *Node) processCheckpoint(req Request, clientID int) {
	// TODO: Add checkpoint support
	// Empty for now
}

func (nd *Node) isInClientLog(req Request) bool {
	req, ok := nd.clientMessageLog[clientMessageLogItem{req.Inner.Id, req.Inner.Timestamp}]
	if !ok {
		return false
	} else {
		return true
	}
}

func (nd *Node) addClientLog(req Request) {
	if !nd.isInClientLog(req) {
		nd.clientMessageLog[clientMessageLogItem{req.Inner.Id, req.Inner.Timestamp}] = req
	}
	return // We don't have logs for now
}

func (nd *Node) handleTimeout(dig DigType, view int) {
	nd.mu.Lock()
	if nd.view > view {
		nd.mu.Unlock()
		return
	}
	MyPrint(2, "%s", "Timeout triggered.\n")
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
	for k, _ := range nd.prepared {
		r, _ := nd.nodeMessageLog.get(typePrePrepare, k, nd.primary) // old primary
		e.Encode(r)
		counter := 0
		for i := 1; i < nd.N; i++ {
			if counter == 2*nd.f {
				rtmp := Request{} // to make sure that alwyas N reqs are encoded
				rtmp.Inner = RequestInner{}
				rtmp.Inner.Id = -1
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

	viewChange := nd.createRequest(typeViewChange, nd.lastStableCheckpoint, MsgType(b.Bytes()))
	nd.broadcast(viewChange) // TODO: broadcast view change RPC path.
	nd.processViewChange(viewChange, 0)
}

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
	MyPrint(2, "[%d] Received request from client: %v.\n", nd.id, req)
	// TODO: do we need to verify client's request?
	nd.mu.Lock()
	if !nd.viewInUse {
		nd.mu.Unlock()
		return
	}
	nd.addClientLog(req)
	reqTimer := time.NewTimer(time.Duration(nd.timeout) * time.Second)
	go func(t *time.Timer, r Request) {
		<-reqTimer.C
		nd.handleTimeout(req.Dig, req.Inner.View)
	}(reqTimer, req)
	nd.active[req.Dig] = ActiveItem{&req, reqTimer, clientID}
	nd.mu.Unlock()

	if nd.primary == nd.id {
		MyPrint(2, "[%d] Leader acked: %v.\n", nd.id, req)
		nd.seq = nd.seq + 1
		m := nd.createRequest(typePrePrepare, nd.seq, MsgType(req.Dig))
		nd.nodeMessageLog.set(m.Inner.Reqtype, m.Inner.Seq, m.Inner.Id, m)
		// write log
		nd.broadcast(m) // TODO: broadcast pre-prepare RPC path.
	}
}

func (nd *Node) initializeKeys() {
	gob.Register(&ecdsa.PrivateKey{})
	// fmt.Println("============")
	// MyPrint(2, string(nd.N))
	// MyPrint(2, string(nd.cfg.N)
	// fmt.Println("============")
	for i := 0; i < nd.N; i++ {
		pubkeyFile := fmt.Sprintf("sign%v.pub", i)
		fmt.Println("fetching file: ", pubkeyFile)
		pubKey := FetchPublicKey(path.Join(nd.cfg.KD, pubkeyFile))
		nd.keyDict[i] = pubKey
		if i == nd.id {
			pemkeyFile := fmt.Sprintf("sign%v.pem", i)
			pemKey := FetchPrivateKey(path.Join(nd.cfg.KD, pemkeyFile))
			nd.ecdsaKey = pemKey
			MyPrint(1, "Fetched private key")
		}
	}
}

func (nd *Node) incPrepDict(dig DigType) {
	if val, ok := nd.prepDict[dig]; ok {
		val.number++
		nd.prepDict[dig] = val
	} else {
		nd.prepDict[dig] = reqCounter{1, false, nil}
	}
}

func (nd *Node) incCommDict(dig DigType) {
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
				val, ok := nd.nodeMessageLog.get(typePrePrepare, seq, nd.primary)
				MyPrint(2, "Dig: %v, val.Dig: %v\n", dig, DigType(val.Inner.Msg))
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
				val, ok := nd.nodeMessageLog.get(typePrePrepare, seq, nd.primary)
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

func (nd *Node) processPrePrepare(req Request, clientID int) {

	seq := req.Inner.Seq
	if val1, ok1 := nd.nodeMessageLog.content[typePrePrepare]; ok1 {
		if _, ok2 := val1[seq]; ok2 {
			return
		}
	}
	// TODO: check client signature

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

	nd.nodeMessageLog.set(req.Inner.Reqtype, req.Inner.Seq, req.Inner.Id, req)
	m := nd.createRequest(typePrepare, req.Inner.Seq, MsgType(req.Inner.Msg)) // TODO: check content!
	nd.nodeMessageLog.set(m.Inner.Reqtype, m.Inner.Seq, m.Inner.Id, m)
	nd.recordPBFT(m)
	nd.mu.Lock()
	nd.incPrepDict(DigType(req.Inner.Msg))
	nd.mu.Unlock()
	nd.broadcast(m)
	if nd.checkPrepareMargin(DigType(req.Inner.Msg), req.Inner.Seq) { // TODO: check dig vs Inner.Msg
		nd.record("PREPARED sequence number " + strconv.Itoa(req.Inner.Seq) + "\n")
		m := nd.createRequest(typeCommit, req.Inner.Seq, req.Inner.Msg) // TODO: check content
		nd.broadcast(m)
		nd.nodeMessageLog.set(m.Inner.Reqtype, m.Inner.Seq, m.Inner.Id, m)
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
	res := fmt.Sprintf("[%d] seq:%d from:%d view:%d %s\n", req.Inner.Reqtype, req.Inner.Seq, req.Inner.Id, req.Inner.View, reqsummary)
	nd.recordCommit(res)
}

func (nd *Node) recordPBFT(req Request) {
	reqsummary := repr.String(&req)
	res := fmt.Sprintf("[%d] seq:%d from:%d view:%d %s\n", req.Inner.Reqtype, req.Inner.Seq, req.Inner.Id, req.Inner.View, reqsummary)
	nd.record(res)
}

func (nd *Node) addNodeHistory(req Request) {
	nd.nodeMessageLog.set(req.Inner.Reqtype, req.Inner.Seq, req.Inner.Id, req)
}

func (nd *Node) processPrepare(req Request, clientID int) {
	nd.addNodeHistory(req)
	nd.mu.Lock()
	nd.incPrepDict(DigType(req.Inner.Msg))
	nd.mu.Unlock()
	if nd.checkPrepareMargin(DigType(req.Inner.Msg), req.Inner.Seq) { // TODO: check dig vs Inner.Msg
		MyPrint(2, "[%d] Checked Margin\n", nd.id)
		nd.record("PREPARED sequence number " + strconv.Itoa(req.Inner.Seq) + "\n")
		m := nd.createRequest(typeCommit, req.Inner.Seq, req.Inner.Msg) // TODO: check content
		nd.broadcast(m)
		nd.nodeMessageLog.set(m.Inner.Reqtype, m.Inner.Seq, m.Inner.Id, m)
		nd.mu.Lock()
		nd.incCommDict(m.Dig) //TODO: check content
		nd.mu.Unlock()
		nd.recordPBFT(m)
		nd.prepared[req.Inner.Seq] = m // or msg?
		if nd.checkCommittedMargin(m.Dig, m) {
			nd.record("COMMITED seq number " + strconv.Itoa(m.Inner.Seq) + "\n")
			nd.recordPBFTCommit(m)
			nd.executeInOrder(m)
		}
	}
}

func (nd *Node) processCommit(req Request) {
	nd.addNodeHistory(req)
	nd.mu.Lock()
	nd.incCommDict(DigType(req.Inner.Msg))
	nd.mu.Unlock()
	if nd.checkCommittedMargin(DigType(req.Inner.Msg), req) {
		MyPrint(2, "[%d] Committed %v\n", nd.id, req)
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
	key := nd.keyDict[req.Inner.Id]
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
			if reqTemp.Inner.Id == v2.Inner.Id {
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
	MyPrint(1, "%s", "Receiveed a view change req from "+strconv.Itoa(req.Inner.Id))
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
	vpreDict := make(map[int]Request)
	vprepDict := make(viewDict)
	m := req.Inner.Msg
	bufm := bytes.Buffer{}
	bufm.Write([]byte(m))
	//// extract msgs in m
	dec := gob.NewDecoder(&bufm)
	var lenCkPf, preparedLen int
	checkpointProofT := make([]Request, 0)
	vpreDict = make(map[int]Request)
	vprepDict = make(viewDict)
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
			if rt.Inner.Id >= 0 { // so that this is not a dummy req
				switch rt.Inner.Reqtype {
				case typePrepare:
					if _, ok := vprepDict[rt.Inner.Seq]; !ok {
						vprepDict[rt.Inner.Seq] = make(map[int]Request)
					}
					vprepDict[rt.Inner.Seq][rt.Inner.Id] = rt
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
			r, _ := nd.nodeMessageLog.get(typePrePrepare, j, nd.primary)
			tmp := nd.createRequest(typePrePrepare, j, r.Inner.Msg)
			reqList = append(reqList, tmp)
			//e.Encode(tmp)
		}
		e.Encode(reqList)
		out := nd.createRequest(typeNewView, 0, MsgType(buf.Bytes()))
		nd.viewInUse = true
		nd.primary = nd.view % nd.N
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
			out := nd.createRequest(typePrepare, r.Inner.Seq, r.Inner.Msg)
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
		MyPrint(3, "%s", "Failed view change")
		return
	}
	if req.Inner.View >= nd.view {
		nd.view = req.Inner.View
		nd.viewInUse = true
		nd.primary = nd.view % nd.N
		nd.active = make(map[DigType]ActiveItem)
		nd.resetMsgDicts()
		nd.clientMessageLog = make(map[clientMessageLogItem]Request)
		nd.prepared = make(map[int]Request)
		nd.newViewProcessPrePrepare(prprList)
		MyPrint(2, "%s", "New View accepted")
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
	MyPrint(1, "[%d] Begin to setup RPC connections.\n", nd.id)
	peers := make([]*rpc.Client, nd.cfg.N)
	for i := 0; i < nd.cfg.N; i++ {
		cl, err := rpc.Dial("tcp", nd.cfg.IPList[i]+":"+strconv.Itoa(nd.cfg.Ports[i]))
		if err != nil {
			MyPrint(3, "RPC error.\n")
		}
		peers[i] = cl
	}
	nd.peers = peers
	MyPrint(1, "[%d] Setup finished\n", nd.id)
}

func Make(cfg Config, me int, port int, view int, applyCh chan ApplyMsg, max_requests int) *Node {
	gob.Register(&ApplyMsg{})
	gob.Register(&RequestInner{})
	gob.Register(&Request{})
	gob.Register(&checkpointProofType{})
	gob.Register(&ProxyProcessPrePrepareArg{})
	gob.Register(&ProxyProcessPrePrepareReply{})
	gob.Register(&ProxyNewClientRequestArg{})
	gob.Register(&ProxyNewClientRequestReply{})
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
	rpc.Register(nd)
	nd.cfg = cfg
	nd.N = cfg.N
	nd.f = (nd.N - 1) / 3
	MyPrint(2, "Going to tolerate %d adversaries\n", nd.f)
	nd.lowBound = 0
	nd.highBound = 0
	nd.view = view
	nd.viewInUse = true
	nd.primary = view % nd.cfg.N
	nd.port = port
	nd.seq = 0
	nd.lastExecuted = 0
	nd.lastStableCheckpoint = 0
	nd.checkpointProof = make([]checkpointProofType, 0)
	nd.timeout = 600
	//nd.mu = sync.Mutex{}
	nd.clientBuffer = ""
	//nd.clientMu = sync.Mutex{}
	nd.keyDict = make(map[int]keyItem)
	nd.checkpointInterval = 100

	nd.vmin = 0
	nd.vmax = 0
	nd.waiting = nil
	nd.max_requests = max_requests
	nd.kill_flag = false
	nd.id = me
	nd.active = make(map[DigType]ActiveItem)
	nd.applyCh = applyCh
	nd.prepDict = make(map[DigType]reqCounter)
	nd.commDict = make(map[DigType]reqCounter)
	nd.nodeMessageLog = nodeMsgLog{}
	nd.nodeMessageLog.content = make(map[int](map[int](map[int]Request)))
	nd.prepared = make(map[int]Request)
	MyPrint(2, "Initial Config %v\n", nd)
	nd.cfg.LD = path.Join(GetCWD(), "logs/")
	// kfpath := path.Join(cfg.KD, filename)

	MakeDirIfNot(nd.cfg.LD) //handles 'already exists'
	fi, err := os.Create(path.Join(nd.cfg.LD, "PBFTLog"+strconv.Itoa(nd.id)+".txt"))
	if err == nil {
		nd.outputLog = fi
	} else {
		panic(err)
	}
	fi2, err2 := os.Create(path.Join(nd.cfg.LD, "PBFTBuffer"+strconv.Itoa(nd.id)+".txt"))
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
	return nd
}
