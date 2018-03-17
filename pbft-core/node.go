package pbft

import "time"
import (
	"fmt"
	"sync"
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
	"net/rpc"
	"github.com/alecthomas/gometalinter/_linters/github.com/client9/misspell"
	"github.com/rogpeppe/godef/go/ast"
	"encoding/base64"
	"encoding/gob"
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"net"
	"strconv"
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

type DigType string

type nodeMsgLog struct {
	content map[int](map[int](map[int]Request))
}

func (nml *nodeMsgLog) get(typ int, seq int, id int) (Request, bool) {
	if val, ok := nml.content[typ]; ok {
		if val2, ok:= val[seq]; ok {
			if val3, ok:=val2[id]; ok {
				return val3, true
			} else {
				return Request{}, false
			}
		}
	}
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

type checkpointProofType []byte

type hellowSignature big.Int

type keyItem []byte

type Node struct {
	mu 				sync.Mutex
	clientMu		sync.Mutex
	peers 			[]*rpc.Client
	port			int
	max_requests	int
	kill_flag		bool

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
	viewDict				map[int]([]int)
	keyDict					map[int]keyItem


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
	req Request
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
	req := Request{m, nil, nil}
	req.inner.outer = &req
	req.addSig(key)
	return req
}

func (req *Request) addSig(privKey *ecdsa.PrivateKey) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(req.inner)
	if err != nil {
		myPrint(3, `failed to encode!`)
	}
	s := getHash(string(b.Bytes()))
	sigr, sigs, err := ecdsa.Sign(rand.Reader, privKey, []byte(s))
	if err != nil {
		myPrint(3, "Error signing.")
		return
	}
	req.dig = s
	req.sig = msgSignature{sigr, sigs}
}


func (nd *Node) resetMsgDicts() {
	nd.nodeMessageLog = nodeMsgLog{}
	nd.prepDict = make(map[DigType]prepareDictItem)
	nd.commDict = make(map[DigType]commitDictItem)
}

func (nd *Node) execute(am ApplyMsg) {
	// TODO: add msg to applyCh, should be executed in a separate go routine
	// TODO: add we probably want to keep a log for this
	nd.applyCh <- am
}

func (nd *Node) suicide() {
	nd.kill_flag = true
}

func (nd *Node) broadcast(req Request) {
	switch req.inner.reqtype {
	case TYPE_PRPR:
		nd.broadcastByRPC("Node.ProcessPrePare", )
		break
	case TYPE_REQU:
		nd.broadcastByRPC("Node.NewClientRequest")
	case TYPE_PREP:
		nd.broadcastByRPC("Node.ProcessPrepare")
	case TYPE_COMM:
		nd.broadcastByRPC("Node.ProcessCommit")
	case TYPE_VCHA:
		nd.broadcastByRPC("Node.ProcessViewChange")
	case TYPE_NEVW:
		nd.broadcastByRPC("Node.ProcessNewView")
	case TYPE_CHKP:
		nd.broadcastByRPC("Node.ProcessCheckpoint")
	default:
		panic(1)  // something bad

	}
}

// broadcast to all the peers
func (nd *Node) broadcastByRPC(rpcPath string, arg interface{}, reply []*interface{}) {
	divCallList := make([]*rpc.Call, 0)
	for ind, c := range nd.peers {
		if ind == nd.id {
			continue  // skip the node itself
		}
		divCallList = append(divCallList, c.Go(rpcPath, arg, reply[ind], nil))
	}
	// synchronize
	for _, divCall := range divCallList {
		divCallDone := <- divCall.Done
		if divCallDone.Error != nil {
			myPrint(3, "error happened in broadcasting " + rpcPath + "\n")
		}
	}
}

func (nd *Node) executeInOrder(req Request) {
	// Not sure if we need this inside the protocol. I tend to put this in application layer.
	waiting := true
	r := req
	seq := req.inner.seq
	dig := req.dig
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

	ln, err := net.Listen("tcp", ":" + strconv.Itoa(nd.port))
	if err != nil {
		myPrint(3, "Error in listening...")
		return
	}
	counter := 0

	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(c)

		// get events from the buffer
		// TODO: finish server loop
	}
}

func (nd *Node) clean() {
	// TODO: do clean work
}

func (nd *Node) ProcessCheckpoint() {
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
	myPrint(2, "Timeout triggered.\n")
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
	for cp := range nd.checkpointProof {
		e.Encode(cp)
	}

	for k, v := range nd.prepared {
		r, _ := nd.nodeMessageLog.get(TYPE_PRPR, k, nd.primary) // old primary
		e.Encode(v)
		counter := 0
		for i := 1; i < nd.N; i++ {
			if counter == 2*nd.f {
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

	viewChange := nd.createRequest(TYPE_VCHA, nd.lastStableCheckpoint, MsgType(b))
	nd.broadcast() // TODO: broadcast view change RPC path.
	nd.ProcessViewChange(viewChange, 0)
}

func (nd *Node) NewClientRequest(req Request, clientId int) {  // TODO: change to single arg and single reply
	if v, ok := nd.active[req.dig]; ok {
		if v.req == nil {
			nd.active[req.dig] = ActiveItem{&req, v.t, v.clientId}
			if v2, ok2 := nd.commDict[req.dig]; ok2 && v2.prepared{
				msg := v2.req
				nd.executeInOrder(msg)
			}
			return
		}
	}

	nd.mu.Lock()
	if !nd.viewInUse {
		nd.mu.Unlock()
		return
	}
	nd.AddClientLog(req)
	reqTimer := time.NewTimer(time.Duration(nd.timeout) * time.Second)
	go func(r Request){
		<- reqTimer.C
		nd.HandleTimeout(req.dig, req.inner.view)
	}(req)
	nd.active[req.dig] = ActiveItem{&req, reqTimer, clientId}
	nd.mu.Unlock()

	if nd.primary == nd.id {
		nd.seq = nd.seq + 1
		m := nd.createRequest(TYPE_PRPR, nd.seq, req.dig)
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
						valt.req = req
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

func (nd *Node) ProcessPrePrepare() {

}

func (nd *Node) ProcessPrepare() {

}

func (nd *Node) ProcessCommit() {

}

func (nd *Node) VirtualProcessCheckPoint() {

}

func (nd *Node) VirtualProcessPrepare() {

}

func (nd *Node) ProcessViewChange() {

}

func (nd *Node) NewViewProcessPrePrepare() {

}

func (nd *Node) NewViewProcessView() {

}

func (nd *Node) ProcessNewView() {

}

func (nd *Node) IsInNodeLog() {

}

func (nd *Node) ProcessRequest() {

}

func Make(peers []*rpc.Client, me int, port int, view int, applyCh chan ApplyMsg, max_requests int) *Node {
	gob.Register(ApplyMsg{})
	gob.Register(RequestInner{})
	gob.Register(Request{})
	gob.Register(checkpointProofType{})
	rpc.Register(Node{})
	nd := &Node{}
	nd.N = len(peers)
	nd.peers = peers
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
	nd.checkpointProof = make([]byte, 0)
	nd.timeout = 600
	//nd.mu = sync.Mutex{}
	nd.clientBuffer = ""
	nd.port =
	//nd.clientMu = sync.Mutex{}


	nd.checkpointInterval = 100
	nd.vmin = 0
	nd.vmax = 0
	nd.waiting = nil
	nd.max_requests = max_requests
	nd.kill_flag = false
	nd.id = me
	nd.applyCh = applyCh
	///// TODO: set up ECDSA

	nd.clientMessageLog = make(map[clientMessageLogItem]Request)
	nd.InitializeKeys()

	go nd.serverLoop()
	return nd
}