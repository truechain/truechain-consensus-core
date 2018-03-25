/*
The MIT License (MIT)

Copyright (c) 2018 TrueChain Foundation

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package pbft

import "time"
import (
	"fmt"
	"sync"
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
	"net/rpc"
	//"github.com/alecthomas/gometalinter/_linters/github.com/client9/misspell"
	//"github.com/rogpeppe/godef/go/ast"
	"encoding/base64"
	"encoding/gob"
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"net"
	"strconv"
	"os"
	"github.com/alecthomas/repr"
	//"golang.org/x/tools/go/gcimporter15/testdata"
	//"log"
)

const SERVER_PORT = 40162
const BUFFER_SIZE = 4096 * 32
const (
	TYPE_PRPR = iota
	TYPE_PREP = iota
	TYPE_COMM = iota
	TYPE_INIT = iota
	TYPE_REQU = iota
	TYPE_VCHA = iota
	TYPE_NEVW = iota
	TYPE_CHKP = iota
)

// TODO: change all the int to int64 in case of overflow

// go binary encoder
func ToGOB64(m []byte) string {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(m)
	if err != nil { fmt.Println(`failed gob Encode`, err) }
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

// go binary decoder
func FromGOB64(str string) []byte {
	m := make([]byte, 0)
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil { fmt.Println(`failed base64 Decode`, err); }
	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	err = d.Decode(&m)
	if err != nil { fmt.Println(`failed gob Decode`, err); }
	return m
}

func getHash(plaintext string) string {
	hasher := sha512.New()
	hasher.Write([]byte(plaintext))
	return hex.EncodeToString(hasher.Sum(nil))
}

type ApplyMsg struct {
	Index	int
	Content interface{}
}

type clientMessageLogItem struct {
	clientID int
	timestamp	int64
}

type viewDictKey struct {
	seq int
	id int
}

type DigType string

type nodeMsgLog struct {
	content map[int](map[int](map[int]Request))
}

func (nml *nodeMsgLog) get(typ int, seq int, id int) (Request, bool) {
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
	_, _ = nml.get(typ, seq, id)  // in case there is an access error
	nml.content[typ][seq][id] = req
}

type commitDictItem struct {
	numberOfCommit	int
	committed bool
	prepared bool
	req Request
}

type checkpointProofType Request

type viewDictItem struct {
	store []Request
	holder1 int
	holder2 int
}

type viewDictType map[int]viewDictItem

type hellowSignature big.Int

type keyItem *ecdsa.PublicKey

type Node struct {
	cfg				Config
	mu 				sync.Mutex
	clientMu		sync.Mutex
	peers 			[]*rpc.Client
	port			int
	max_requests	int
	kill_flag		bool

	ListenReady		chan bool

	ecdsaKey		*ecdsa.PrivateKey
	helloSignature	*hellowSignature
	connections 	int
	id				int
	N 				int
	view			int
	viewInUse		bool
	f 				int
	lowBound 		int
	highBound 		int
	primary 		int
	seq 			int
	lastExecuted 	int
	lastStableCheckpoint 	int
	checkpointProof 		[]checkpointProofType
	checkpointInterval 		int
	vmin					int
	vmax 					int
	waiting 				map[int]Request
	timeout 				int
	clientBuffer 			string
	active 					map[DigType]ActiveItem
	prepared				map[int]Request
	prepDict				map[DigType]reqCounter
	commDict				map[DigType]reqCounter
	viewDict				viewDictType
	keyDict					map[int]keyItem

	outputLog				*os.File
	commitLog				*os.File

	nodeMessageLog			nodeMsgLog

	/// log
	clientMessageLog		map[clientMessageLogItem]Request
	// nodeMessageLog	map[string](map[int](map[int]Request))

	/// apply channel

	applyCh					chan ApplyMsg   // Make sure that the client keeps getting the content out of applyCh
}

type reqCounter struct {
	number int
	prepared bool
	req *Request
}

type Log struct {
	//TODO
}

type ActiveItem struct {
	req *Request
	t	*time.Timer
	clientId int
}

type MsgType string

type RequestInner struct {
	id int
	seq int
	view int
	reqtype int  //or int?
	msg MsgType
	timestamp int64
	outer *Request
}

type msgSignature struct {
	r *big.Int
	s *big.Int
}

type Request struct {
	inner RequestInner
	dig DigType
	sig msgSignature
}

func (nd *Node) createRequest(reqType int, seq int, msg MsgType) Request {
	key := nd.ecdsaKey
	m := RequestInner{nd.id, seq, nd.view, reqType, msg, -1, nil}
	req := Request{m, "", msgSignature{nil, nil}}
	req.inner.outer = &req
	req.addSig(key)
	return req
}

func (req *Request) addSig(privKey *ecdsa.PrivateKey) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(req.inner)
	if err != nil {
		myPrint(3, "%s", `failed to encode!`)
	}
	s := getHash(string(b.Bytes()))
	sigr, sigs, err := ecdsa.Sign(rand.Reader, privKey, []byte(s))
	if err != nil {
		myPrint(3, "%s", "Error signing.")
		return
	}
	req.dig = DigType(s)
	req.sig = msgSignature{sigr, sigs}
}

func (req *Request) verifyDig() bool {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(req.inner)
	if err != nil {
		myPrint(3, "%s", `failed to encode!`)
	}
	s := getHash(string(b.Bytes()))
	return s == string(req.dig)
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
	req Request
	clientID int
}

type ProxyProcessPrePrepareReply struct {

}

type ProxyNewClientRequestArg struct {
	req Request
	clientID int
}

type ProxyNewClientRequestReply struct {

}

type ProxyProcessPrepareArg struct {
	req Request
	clientID int
}

type ProxyProcessPrepareReply struct {

}

type ProxyProcessCommitArg struct {
	req Request
}

type ProxyProcessCommitReply struct {

}

type ProxyProcessViewChangeArg struct {
	req Request
	from int
}

type ProxyProcessViewChangeReply struct {

}

type ProxyProcessNewViewArg struct {
	req Request
	clientId int
}

type ProxyProcessNewViewReply struct {

}

type ProxyProcessCheckpointArg struct {
	req Request
	clientId int
}

type ProxyProcessCheckpointReply struct {

}


func (nd *Node) broadcast(req Request) {

	switch req.inner.reqtype {
	// refine the following
	case TYPE_PRPR:
		arg := ProxyProcessPrePrepareArg{req, req.inner.id}
		reply := make([]interface{}, nd.N)
		for k:= 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessPrePrepareReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessPrePrepare", arg, &reply)
		break
	case TYPE_REQU:
		arg := ProxyNewClientRequestArg{req, req.inner.id}
		reply := make([]interface{}, nd.N)
		for k:= 0; k < nd.N; k++ {
			reply[k] = &ProxyNewClientRequestReply{}
		}
		nd.broadcastByRPC("Node.ProxyNewClientRequest", arg, &reply)
		break
	case TYPE_PREP:
		arg := ProxyProcessPrepareArg{req, req.inner.id}
		reply := make([]interface{}, nd.N)
		for k:= 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessPrepareReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessPrepare", arg, &reply)
		break
	case TYPE_COMM:
		arg := ProxyProcessCommitArg{req}
		reply := make([]interface{}, nd.N)
		for k:= 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessCommitReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessCommit", arg, &reply)
		break
	case TYPE_VCHA:
		arg := ProxyProcessViewChangeArg{req, nd.id}  // a dummy from
		reply := make([]interface{}, nd.N)
		for k:= 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessViewChangeReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessViewChange", arg, &reply)
		break
	case TYPE_NEVW:
		arg := ProxyProcessNewViewArg{req, req.inner.id}
		reply := make([]interface{}, nd.N)
		for k:= 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessNewViewReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessNewView", arg, &reply)
		break
	case TYPE_CHKP:
		arg := ProxyProcessCheckpointArg{req, req.inner.id}
		reply := make([]interface{}, nd.N)
		for k:= 0; k < nd.N; k++ {
			reply[k] = &ProxyProcessCheckpointReply{}
		}
		nd.broadcastByRPC("Node.ProxyProcessCheckpoint", arg, &reply)
		break
	default:
		panic(1)  // something bad

	}
}

func (nd *Node) ProxyProcessPrePrepare(arg ProxyProcessPrePrepareArg, reply *ProxyProcessPrePrepareReply) error {
	nd.ProcessPrePrepare(arg.req, arg.clientID)  // we don't have return value here
	return nil
}

func (nd *Node) ProxyNewClientRequest(arg ProxyNewClientRequestArg, reply *ProxyNewClientRequestReply) error {
	nd.NewClientRequest(arg.req, arg.clientID)  // we don't have return value here
	return nil
}

func (nd *Node) ProxyProcessPrepare(arg ProxyProcessPrepareArg, reply *ProxyProcessPrepareReply) error {
	nd.ProcessPrepare(arg.req, arg.clientID)  // we don't have return value here
	return nil
}

func (nd *Node) ProxyProcessCommit(arg ProxyProcessCommitArg, reply *ProxyProcessCommitReply) error {
	nd.ProcessCommit(arg.req)  // we don't have return value here
	return nil
}

func (nd *Node) ProxyProcessViewChange(arg ProxyProcessViewChangeArg, reply *ProxyProcessViewChangeReply) error {
	nd.ProcessViewChange(arg.req, arg.from)
	return nil
}

func (nd *Node) ProxyProcessNewView(arg ProxyProcessNewViewArg, reply *ProxyProcessNewViewReply) error {
	nd.ProcessNewView(arg.req, arg.clientId)
	return nil
}

func (nd *Node) ProxyProcessCheckpoint(arg ProxyProcessCheckpointArg, reply *ProxyProcessCheckpointReply) error {
	nd.ProcessCheckpoint(arg.req, arg.clientId)
	return nil
}

// broadcast to all the peers
func (nd *Node) broadcastByRPC(rpcPath string, arg interface{}, reply *[]interface{}) {
	divCallList := make([]*rpc.Call, 0)
	for ind, c := range nd.peers {
		if ind == nd.id {
			continue  // skip the node itself
		}
		divCallList = append(divCallList, c.Go(rpcPath, arg, (*reply)[ind], nil))
	}
	// synchronize
	for _, divCall := range divCallList {
		divCallDone := <- divCall.Done
		if divCallDone.Error != nil {
			myPrint(3, "%s", "error happened in broadcasting " + rpcPath + "\n")
		}
	}
}

func (nd *Node) executeInOrder(req Request) {
	// Not sure if we need this inside the protocol. I tend to put this in application layer.
	waiting := true
	r := req
	seq := req.inner.seq
	dig := DigType(req.inner.msg)
	ac := nd.active[dig]
	if ac.req == nil {
		return
	}
	for seq == nd.lastExecuted + 1 {
		waiting = false
		nd.execute(ApplyMsg{seq, r})
		nd.lastExecuted += 1
		rtemp, ok := nd.waiting[seq + 1]
		if ok {
			seq += 1
			r = rtemp
			delete(nd.waiting, seq)
		}

	}

	if waiting {
		nd.waiting[req.inner.seq] = req
	}
}

func (nd *Node) serverLoop() {
	addy, err := net.ResolveTCPAddr("tcp", "0.0.0.0:" + strconv.Itoa(nd.port))
	if err != nil {
		myPrint(3,  "Error in resolving tcp addr...\n")
		return
	}

	ln, err := net.ListenTCP("tcp", addy)
	if err != nil {
		myPrint(3,  "Error in listening...\n")
		return
	}
	//counter := 0
	// TODO: add timer to try client

	rpc.Register(Node{})
	rpc.Accept(ln)

	nd.ListenReady <- true  // trigger the connection
}

func (nd *Node) clean() {
	// TODO: do clean work
}

func (nd *Node) ProcessCheckpoint(req Request, clientID int) {
	// TODO: Add checkpoint support
	// Empty for now
}

func (nd *Node) IsInClientLog(req Request) bool {
	req, ok := nd.clientMessageLog[clientMessageLogItem{req.inner.id, req.inner.timestamp}]
	if !ok {
		return false
	} else {
		return true
	}
}

func (nd *Node) AddClientLog(req Request) {
	if !nd.IsInClientLog(req) {
		nd.clientMessageLog[clientMessageLogItem{req.inner.id, req.inner.timestamp}]  = req
	}
	return // We don't have logs for now
}

func (nd *Node) HandleTimeout(dig DigType, view int) {
	nd.mu.Lock()
	if nd.view > view {
		nd.mu.Unlock()
		return
	}
	myPrint(2, "%s", "Timeout triggered.\n")
	for k, v := range nd.active {
		if k != dig {
			v.t.Stop()
		}
	}
	nd.view += 1
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
		r, _ := nd.nodeMessageLog.get(TYPE_PRPR, k, nd.primary) // old primary
		e.Encode(r)
		counter := 0
		for i := 1; i < nd.N; i++ {
			if counter == 2*nd.f {
				rtmp := Request{}  // to make sure that alwyas N reqs are encoded
				rtmp.inner = RequestInner{}
				rtmp.inner.id = -1
				e.Encode(rtmp)
				break
			}
			r, ok := nd.nodeMessageLog.get(TYPE_PREP, k, i)
			if ok {
				e.Encode(r)
				counter += 1
				// TODO: add counter whenever we found a PREP
			}
		}
	}

	viewChange := nd.createRequest(TYPE_VCHA, nd.lastStableCheckpoint, MsgType(b.Bytes()))
	nd.broadcast(viewChange) // TODO: broadcast view change RPC path.
	nd.ProcessViewChange(viewChange, 0)
}

func (nd *Node) NewClientRequest(req Request, clientId int) {  // TODO: change to single arg and single reply
	if v, ok := nd.active[req.dig]; ok {
		if v.req == nil {
			nd.active[req.dig] = ActiveItem{&req, v.t, v.clientId}
			if v2, ok2 := nd.commDict[req.dig]; ok2 && v2.prepared{
				msg := v2.req
				nd.executeInOrder(*msg)
			}
			return
		}
	}
	// TODO: do we need to verify client's request?
	nd.mu.Lock()
	if !nd.viewInUse {
		nd.mu.Unlock()
		return
	}
	nd.AddClientLog(req)
	reqTimer := time.NewTimer(time.Duration(nd.timeout) * time.Second)
	go func(t *time.Timer, r Request){
		<- reqTimer.C
		nd.HandleTimeout(req.dig, req.inner.view)
	}(reqTimer, req)
	nd.active[req.dig] = ActiveItem{&req, reqTimer, clientId}
	nd.mu.Unlock()

	if nd.primary == nd.id {
		nd.seq = nd.seq + 1
		m := nd.createRequest(TYPE_PRPR, nd.seq, MsgType(req.dig))
		nd.nodeMessageLog.set(m.inner.reqtype, m.inner.seq, m.inner.id, m)
		// write log
		nd.broadcast(m)  // TODO: broadcast pre-prepare RPC path.
	}
}

func (nd *Node) InitializeKeys() {
	// TODO: initialize ECDSA keys and hello signature
	nd.ecdsaKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)  // TODO: read from file

}

func (nd *Node) IncPrepDict(dig DigType) {
	if val, ok := nd.prepDict[dig]; ok {
		val.number += 1
		nd.prepDict[dig] = val
	} else {
		nd.prepDict[dig] = reqCounter{1, false, nil}
	}

}

func (nd *Node) IncCommDict(dig DigType) {
	if val, ok := nd.commDict[dig]; ok {
		val.number += 1
		nd.commDict[dig] = val
	} else {
		nd.commDict[dig] = reqCounter{1, false, nil}
	}
}

func (nd *Node) CheckPrepareMargin(dig DigType, seq int) bool {
	if val, ok := nd.prepDict[dig]; ok {
		if !val.prepared {
			if val.number >= 2 * nd.f + 1 {
				val, ok := nd.nodeMessageLog.get(TYPE_PRPR, seq, nd.primary)
				if ok && val.dig == dig {   // TODO: check the diff!
				//if ok && val.inner.msg == dig {
					if valt, okt := nd.prepDict[dig]; okt {
						valt.prepared = true
						nd.prepDict[dig] = valt
						return true
					}
				}
			}
		}
		return false
	} else {
		return false
	}

}

func (nd *Node) CheckCommittedMargin(dig DigType, req Request) bool {
	seq := req.inner.seq
	if val, ok := nd.commDict[dig]; ok {
		if !val.prepared {
			if val.number >= 2 * nd.f + 1 {
				val, ok := nd.nodeMessageLog.get(TYPE_PRPR, seq, nd.primary)
				if ok && val.dig == dig {   // TODO: check the diff!
					//if ok && val.inner.msg == dig {
					if valt, okt := nd.commDict[dig]; okt {
						valt.prepared = true
						valt.req = &req
						nd.commDict[dig] = valt
						return true
					}
				}
			}
		}
		return false
	} else {
		return false
	}

}

func (nd *Node) ProcessPrePrepare(req Request, clientId int) {
	seq := req.inner.seq
	if val1, ok1 := nd.nodeMessageLog.content[TYPE_PRPR]; ok1 {
		if _, ok2 := val1[seq]; ok2 {
			return
		}
	}
	// TODO: check client signature

	if _, ok := nd.active[DigType(req.inner.msg)]; !ok {
		reqTimer := time.NewTimer(time.Duration(nd.timeout) * time.Second)
		go func(t *time.Timer, r Request){
			<- reqTimer.C
			nd.HandleTimeout(DigType(req.inner.msg), req.inner.view)
		}(reqTimer, req)
		nd.active[DigType(req.inner.msg)] = ActiveItem{&req, reqTimer, clientId}
	}

	nd.nodeMessageLog.set(req.inner.reqtype, req.inner.seq, req.inner.id, req)
	m := nd.createRequest(TYPE_PREP, req.inner.seq, MsgType(req.inner.msg))  // TODO: check content!
	nd.nodeMessageLog.set(m.inner.reqtype, m.inner.seq, m.inner.id, m)
	nd.recordPBFT(m)
	nd.IncPrepDict(DigType(req.inner.msg))
	nd.broadcast(m)
	if nd.CheckPrepareMargin(DigType(req.inner.msg), req.inner.seq) {  // TODO: check dig vs inner.msg
		nd.record("PREPARED sequence number " + strconv.Itoa(req.inner.seq) + "\n")
		m := nd.createRequest(TYPE_COMM, req.inner.seq, req.inner.msg) // TODO: check content
		nd.broadcast(m)
		nd.nodeMessageLog.set(m.inner.reqtype, m.inner.seq, m.inner.id, m)
		nd.IncCommDict(DigType(req.inner.msg)) //TODO: check content
		nd.recordPBFT(m)
		nd.prepared[req.inner.seq] = m // or msg?
		if nd.CheckCommittedMargin(DigType(req.inner.msg), m) {
			nd.record("COMMITED seq number " + strconv.Itoa(m.inner.seq) + "\n")
			nd.recordPBFTCommit(m)
			nd.executeInOrder(m)
		}
	}
}

func (nd *Node) recordCommit(content string) {
	if _, err:= nd.commitLog.Write([]byte(content)); err != nil {
		panic(err)
	}
}

func (nd *Node) record(content string) {
	if _, err:= nd.outputLog.Write([]byte(content)); err != nil {
		panic(err)
	}
}

func (nd *Node) recordPBFTCommit(req Request) {
	reqsummary := repr.String(&req)
	res := fmt.Sprintf("[%d] seq:%d from:%d view:%d %s\n", req.inner.reqtype, req.inner.seq, req.inner.id, req.inner.view, reqsummary)
	nd.recordCommit(res)
}


func (nd *Node) recordPBFT(req Request) {
	reqsummary := repr.String(&req)
	res := fmt.Sprintf("[%d] seq:%d from:%d view:%d %s\n", req.inner.reqtype, req.inner.seq, req.inner.id, req.inner.view, reqsummary)
	nd.record(res)
}

func (nd *Node) addNodeHistory(req Request) {
	nd.nodeMessageLog.set(req.inner.reqtype, req.inner.seq, req.inner.id, req)
}

func (nd *Node) ProcessPrepare(req Request, clientId int) {
	nd.addNodeHistory(req)
	nd.IncPrepDict(DigType(req.inner.msg))
	if nd.CheckPrepareMargin(DigType(req.inner.msg), req.inner.seq) {  // TODO: check dig vs inner.msg
		nd.record("PREPARED sequence number " + strconv.Itoa(req.inner.seq) + "\n")
		m := nd.createRequest(TYPE_COMM, req.inner.seq, req.inner.msg) // TODO: check content
		nd.broadcast(m)
		nd.nodeMessageLog.set(m.inner.reqtype, m.inner.seq, m.inner.id, m)
		nd.IncCommDict(m.dig) //TODO: check content
		nd.recordPBFT(m)
		nd.prepared[req.inner.seq] = m // or msg?
		if nd.CheckCommittedMargin(m.dig, m) {
			nd.record("COMMITED seq number " + strconv.Itoa(m.inner.seq) + "\n")
			nd.recordPBFTCommit(m)
			nd.executeInOrder(m)
		}
	}
}

func (nd *Node) ProcessCommit(req Request) {
	nd.addNodeHistory(req)
	nd.IncCommDict(DigType(req.inner.msg))
	if nd.CheckCommittedMargin(DigType(req.inner.msg), req) {
		nd.record("COMMITTED seq number " + strconv.Itoa(req.inner.seq))
		nd.recordPBFTCommit(req)
		nd.executeInOrder(req)
	}
}

func (nd *Node) ViewProcessCheckPoint(vchecklist *[]Request, lastCheckPoint int) bool {
	if lastCheckPoint == 0 {
		return true
	}
	if len(*vchecklist) <= 2 * nd.f {
		return false
	}
	dig := (*vchecklist)[0].dig
	for _, c := range *vchecklist {
		if c.inner.seq != lastCheckPoint || c.dig != dig {
			return false
		}
	}
	return true
}

type viewDict map[int](map[int]Request)  //map[viewDictKey]Request

func (nd *Node) VerifyMsg(req Request) bool {
	key := nd.keyDict[req.inner.id]
	// check the digest
	dv := req.verifyDig()
	// check the signature
	sc := ecdsa.Verify(key, []byte(req.dig), req.sig.r, req.sig.s)
	return dv && sc
}

func (nd *Node) ViewProcessPrepare(vPrepDict viewDict, vPreDict map[int]Request, lastCheckPoint int) (bool, int) {
	max:=0
	counter := make(map[int]int)
	for k1, v1 := range vPrepDict {
		if _, ok := vPreDict[k1]; !ok {
			return false, 0
		}
		reqTemp:= vPreDict[k1]
		dig := reqTemp.dig
		if !nd.VerifyMsg(reqTemp) {
			return false, 0
		}
		for _, v2 := range v1 {
			if !nd.VerifyMsg(v2) {
				return false, 0
			}
			if v2.dig != dig {
				return false, 0
			}
			if reqTemp.inner.id == v2.inner.id {
				return false, 0 // cannot be sent from the same guy
			}
			if v2.inner.seq < lastCheckPoint {
				return false, 0
			}
			if v2.inner.seq > max {
				max = v2.inner.seq
			}
			if _, ok3 := counter[v2.inner.seq]; !ok3 {
				counter[v2.inner.seq]  = 1
			} else {
				counter[v2.inner.seq] += 1
			}
		}
	}
	for k, v := range counter {
		if v < 2 * nd.f {
			return false, 0
		}
		nd.addNodeHistory(vPreDict[k])
	}
	return true, max
}

func (nd *Node) ProcessViewChange(req Request, from int) {
	myPrint(1, "%s", "Receiveed a view change req from " + strconv.Itoa(req.inner.id))
	nd.addNodeHistory(req)
	newV := req.inner.view
	if nd.view != req.inner.view || newV < nd.view {
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
	m := req.inner.msg
	bufm := bytes.Buffer{}
	bufm.Write([]byte(m))
	//// extract msgs in m
	dec := gob.NewDecoder(&bufm)
	var lenCkPf, preparedLen int
	checkpointProofT := make([]Request, 0)
	vpreDict = make(map[int]Request)
	vprepDict = make(viewDict)
	dec.Decode(&lenCkPf)

	for i:=0; i< lenCkPf; i++ {
		ckpf := checkpointProofType{}
		dec.Decode(&ckpf)
		checkpointProofT = append(checkpointProofT, Request(ckpf))
	}
	dec.Decode(&preparedLen)
	for j:=0; j< preparedLen; j++ {
		r:=Request{}
		dec.Decode(&r)
		for k:=0 ; k< nd.N; k++ {
			var rt Request
			dec.Decode(&rt)
			if rt.inner.id >= 0 {  // so that this is not a dummy req
				switch rt.inner.reqtype {
				case TYPE_PREP:
					if _, ok := vprepDict[rt.inner.seq]; !ok {
						vprepDict[rt.inner.seq] = make(map[int]Request)
					}
					vprepDict[rt.inner.seq][rt.inner.id] = rt
					break
				case TYPE_PRPR:
					vpreDict[rt.inner.seq] = rt
				}

			}
		}
	}
	rc1 := nd.ViewProcessCheckPoint(&checkpointProofT, req.inner.seq)  // fix the type of checkPointProofT
	rc2, maxm := nd.ViewProcessPrepare(vprepDict, vpreDict, req.inner.seq)
	if rc1 && rc2 {
		if _, ok := nd.viewDict[newV] ; !ok {
			reqList := make([]Request, 1)
			reqList[0] = req
			nd.viewDict[newV] = viewDictItem{reqList,0,0}
		} else {
			vTemp := nd.viewDict[newV]
			vTemp.store = append(vTemp.store, req)
			nd.viewDict[newV] = vTemp
		}
	}
	if nd.viewDict[newV].holder1 < req.inner.seq {
		tmp:=nd.viewDict[newV]
		nd.viewDict[newV] = viewDictItem{tmp.store, req.inner.seq, tmp.holder2}
	}

	if nd.viewDict[newV].holder2 < maxm {
		tmp:=nd.viewDict[newV]
		nd.viewDict[newV] = viewDictItem{tmp.store, tmp.holder1, maxm}
	}
	if (!nd.viewInUse || newV > nd.view) && len(nd.viewDict[newV].store) > 2*nd.f {
		// process and send the view req
		buf := bytes.Buffer{}
		e := gob.NewEncoder(&buf)
		reqList := make([]Request, 0)
		for i:=range nd.viewDict[newV].store {
			//e.Encode(nd.viewDict[newV][0][i])
			reqList = append(reqList, nd.viewDict[newV].store[i])
		}
		for j:=nd.viewDict[newV].holder1; j<nd.viewDict[newV].holder2; j++ {
			if j == 0 {
				continue
			}
			r, _ := nd.nodeMessageLog.get(TYPE_PRPR, j, nd.primary)
			tmp := nd.createRequest(TYPE_PRPR, j, r.inner.msg)
			reqList = append(reqList, tmp)
			//e.Encode(tmp)
		}
		e.Encode(reqList)
		out := nd.createRequest(TYPE_NEVW, 0, MsgType(buf.Bytes()))
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

func (nd *Node) NewViewProcessPrePrepare(prprList []Request) bool {
	for _, r := range prprList {
		if nd.VerifyMsg(r) && r.verifyDig() {
			out := nd.createRequest(TYPE_PREP, r.inner.seq, r.inner.msg)
			nd.broadcast(out)
		} else {
			return false
		}
	}
	return true
}

func (nd *Node) NewViewProcessView(vchangeList []Request) bool {
	for _, r := range vchangeList {
		if !(nd.VerifyMsg(r) && r.verifyDig()) {
			return false
		}
	}
	return true
}

func (nd *Node) ProcessNewView(req Request, clientID int) {
	m := req.inner.msg
	bufcurrent := bytes.Buffer{}
	bufcurrent.Write([]byte(m))
	e := gob.NewDecoder(&bufcurrent)
	vchangeList := make([]Request, 0)
	prprList := make([]Request, 0)
	//counter := 0
	reqList := make([]Request, 0)
	e.Decode(&reqList)
	for _, r := range reqList {
		if r.verifyDig() && nd.VerifyMsg(r) {
			switch r.inner.reqtype {
			case TYPE_VCHA:
				vchangeList = append(vchangeList, r)
				break
			case TYPE_PRPR:
				prprList = append(prprList, r)
				break
			}
		}
	}
	if !nd.NewViewProcessView(vchangeList) {
		myPrint(3,"%s",  "Failed view change")
		return
	}
	if req.inner.view >= nd.view {
		nd.view = req.inner.view
		nd.viewInUse = true
		nd.primary = nd.view % nd.N
		nd.active = make(map[DigType]ActiveItem)
		nd.resetMsgDicts()
		nd.clientMessageLog = make(map[clientMessageLogItem]Request)
		nd.prepared = make(map[int]Request)
		nd.NewViewProcessPrePrepare(prprList)
		myPrint(2, "%s", "New View accepted")
	}
}

func (nd *Node) IsInNodeLog() bool {
// deprecated
	return false
}

func (nd *Node) ProcessRequest() {
// we use RPC so we don't need a central msg receiver
}

func (nd *Node) BeforeShutdown() {
	nd.outputLog.Close()
}

func (nd *Node) SetupConnections() {
	peers := make([]*rpc.Client, nd.cfg.N)
	for i:= 0; i<nd.cfg.N; i++ {
		cl, err := rpc.Dial("tcp", nd.cfg.IPList[i] + ":" + strconv.Itoa(nd.cfg.Ports[i]))
		if err != nil {
			myPrint(3, "RPC error.\n")
		}
		peers[i] = cl
	}
	nd.peers = peers
}

func Make(cfg Config, me int, port int, view int, applyCh chan ApplyMsg, max_requests int) *Node {
	gob.Register(ApplyMsg{})
	gob.Register(RequestInner{})
	gob.Register(Request{})
	gob.Register(checkpointProofType{})
	gob.Register(ProxyProcessPrePrepareArg{})
	gob.Register(ProxyProcessPrePrepareReply{})
	gob.Register(ProxyNewClientRequestArg{})
	gob.Register(ProxyNewClientRequestReply{})
	gob.Register(ProxyProcessCheckpointArg{})
	gob.Register(ProxyProcessCheckpointReply{})
	gob.Register(ProxyProcessCommitArg{})
	gob.Register(ProxyProcessCommitReply{})
	gob.Register(ProxyProcessNewViewArg{})
	gob.Register(ProxyProcessNewViewReply{})
	gob.Register(ProxyProcessPrepareArg{})
	gob.Register(ProxyProcessPrepareReply{})
	gob.Register(ProxyProcessViewChangeArg{})
	gob.Register(ProxyProcessViewChangeReply{})
	rpc.Register(Node{})

	nd := &Node{}
	nd.cfg = cfg
	//nd.N = len(peers)
	//nd.peers = peers
	nd.f = (nd.N - 1) / 3
	nd.lowBound = 0
	nd.highBound = 0
	nd.view = view
	nd.viewInUse = true
	nd.primary = view % nd.N
	nd.port = port
	nd.seq = 0
	nd.lastExecuted = 0
	nd.lastStableCheckpoint = 0
	nd.checkpointProof = make([]checkpointProofType, 0)
	nd.timeout = 600
	//nd.mu = sync.Mutex{}
	nd.clientBuffer = ""
	//nd.clientMu = sync.Mutex{}


	nd.checkpointInterval = 100
	nd.vmin = 0
	nd.vmax = 0
	nd.waiting = nil
	nd.max_requests = max_requests
	nd.kill_flag = false
	nd.id = me
	nd.applyCh = applyCh

	fi, err := os.Create("pbftlog" + strconv.Itoa(nd.id) + ".txt")
	if err != nil {
		nd.outputLog = fi
	}
	fi2, err2 := os.Create("pbftcommit" + strconv.Itoa(nd.id) + ".txt")
	if err2 != nil {
		nd.commitLog = fi2
	}
	///// TODO: set up ECDSA

	nd.clientMessageLog = make(map[clientMessageLogItem]Request)
	nd.InitializeKeys()
	nd.ListenReady = make(chan bool)

	go nd.serverLoop()
	return nd
}
