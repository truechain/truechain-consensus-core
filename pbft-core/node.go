package pbft

import "time"
import "math/rand"
import (
	"fmt"
	"sync"
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
	"net/rpc"
	"github.com/alecthomas/gometalinter/_linters/github.com/client9/misspell"
	"github.com/rogpeppe/godef/go/ast"
)

const BUFFER_SIZE = 4096 * 32


// TODO: change all the int to int64 in case of overflow

type ApplyMsg struct {
	Index	int
	Content interface{}
}

type clientMessageLogItem struct {
	clientID int
	timestamp	int64
}

type DigType string

type Node struct {
	mu 				sync.Mutex
	clientMu		sync.Mutex
	peers 			[]*rpc.Client
	max_requests	int
	kill_flag		bool
	ecdsaKey		*ecdsa.PrivateKey
	helloSignature	*big.Int
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
	checkpointProof 		[]byte
	checkpointInterval 		int
	vmin					int
	vmax 					int
	waiting 				map[int]Request
	timeout 				int
	clientBuffer 			string
	active 					map[DigType]ActiveItem
	prepared				map[int]Request

	/// log
	clientMessageLog		map[clientMessageLogItem]Request
	// nodeMessageLog	map[string](map[int](map[int]Request))

	/// apply channel

	applyCh					chan ApplyMsg   // Make sure that the client keeps getting the content out of applyCh
}

type Log struct {
	//TODO
}

type ActiveItem struct {
	req *Request
	t	*time.Timer
	clientId int
}

type Request struct {
	id int
	seq int
	view int
	reqtype string  //or int?
	msg interface{}  // any kind of msg
	timestamp int64
	dig DigType
	sig string
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
	nd.waiting = nils
	nd.max_requests = max_requests
	nd.kill_flag = false
	nd.id = me
	nd.applyCh = applyCh
	///// TODO: set up ECDSA

	nd.clientMessageLog = make(map[clientMessageLogItem]Request)

	go nd.serverLoop()
	return nd
}