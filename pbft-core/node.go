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
)

const BUFFER_SIZE = 4096 * 32

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

func (nml *nodeMsgLog) get(typ int, seq int, id int) Request {
	if val, ok := nml.content[typ]; ok {
		if val2, ok:= val[seq]; ok {
			if val3, ok:=val2[id]; ok {
				return val3
			}
		}
	}
}

func (nml *nodeMsgLog) set(typ int, seq int, id int, req Request) {
	_ = nml.get(typ, seq, id)  // in case there is an access error
	nml.content[typ][seq][id] = req
}


type prepareDictItem struct {
	numberOfPrepared int
	prepared bool
}

type commitDictItem struct {
	numberOfCommit	int
	committed bool
}

type checkpointProofType []byte

type hellowSignature big.Int

type keyItem []byte

type Node struct {
	mu 				sync.Mutex
	clientMu		sync.Mutex
	peers 			[]*rpc.Client
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
	prepDict				map[DigType]prepareDictItem
	commDict				map[DigType]commitDictItem
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
	reqtype string  //or int?
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


func (req *Request) addSig(privKey ecdsa.PrivateKey) (*big.Int, *big.Int) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(req.inner)
	if err != nil {
		myPrint(3, `failed to encode!`)
	}
	s := []byte(getHash(string(b.Bytes())))
	r, s, err := ecdsa.Sign(rand.Reader, privKey, s)
	if err != nil {
		myPrint(3, "Error signing.")
		return nil, nil
	}
	req.sig = msgSignature{r, s}
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

// broadcast to all the peers
func (nd *Node) broadcast(rpcPath string, arg interface{}, reply []*interface{}) {
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
	seq := req.seq
	dig := req.dig
	ac := nd.active[dig]
	if ac.req == nil {
		return
	}
	for seq == nd.lastExecuted + 1 {
		waiting = false
		nd.execute(ApplyMsg{seq, r.msg})
		nd.lastExecuted += 1
		rtemp, ok := nd.waiting[seq + 1]
		if ok {
			seq += 1
			r = rtemp
			delete(nd.waiting, seq)
		}

	}

	if waiting {
		nd.waiting[req.seq] = req
	}
}

func (nd *Node) serverLoop() {
	counter := 0

	for {
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
	req, ok := nd.clientMessageLog[clientMessageLogItem{req.id, req.timestamp}]
	if !ok {
		return false
	} else {
		return true
	}
}

func (nd *Node) AddClientLog(req Request) {
	if !nd.IsInClientLog(req) {
		nd.clientMessageLog[clientMessageLogItem{req.id, req.timestamp}]  = req
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

	msg := ""
	for cp := range nd.checkpointProof {
		msg += cp  // type of checkpoint proofs
	}

	for k, v := range nd.prepared {
		msg += v // TODO: serialize the corresponnding Pre-Prepare
		counter := 0
		for i := 1; i < nd.N; i++ {
			if counter == 2*nd.f {
				break
			}
			// TODO: add counter whenever we found a PREP
		}

	}

}

func (nd *Node) NewClientRequest(req Request) {
	if v, ok := nd.active[req.dig]; ok {
		if v.req == nil {

		}
	}

}

func (nd *Node) InitializeKeys() {
	// TODO: initialize ECDSA keys and hello signature
	nd.ecdsaKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)  // TODO: read from file

}

func (nd *Node) IncPrepDict() {

}

func (nd *Node) IncCommDict() {

}

func (nd *Node) CheckPrepareMargin() {

}

func (nd *Node) CheckCommittedMargin() {

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

func (nd *Node) AddNodeLog() {

}

func (nd *Node) IsInNodeLog() {

}

func (nd *Node) ProcessRequest() {

}

func Make(peers []*rpc.Client, me int, view int, applyCh chan ApplyMsg, max_requests int) *Node {
	gob.Register(ApplyMsg{})
	gob.Register(RequestInner{})
	nd := &Node{}
	nd.N = len(peers)
	nd.peers = peers
	nd.f = (nd.N - 1) / 3
	nd.lowBound = 0
	nd.highBound = 0
	nd.view = view
	nd.viewInUse = true
	nd.primary = view % nd.N
	nd.seq = 0
	nd.lastExecuted = 0
	nd.lastStableCheckpoint = 0
	nd.checkpointProof = make([]byte, 0)
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
	///// TODO: set up ECDSA

	nd.clientMessageLog = make(map[clientMessageLogItem]Request)
	nd.InitializeKeys()

	go nd.serverLoop()
	return nd
}