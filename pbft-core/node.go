package pbft

import "time"
import "math/rand"
import (
	"fmt"
	"sync"
	"net/rpc"
	"github.com/alecthomas/gometalinter/_linters/github.com/client9/misspell"
)

const BUFFER_SIZE = 4096 * 32


// TODO: change all the int to int64 in case of overflow

type ApplyMsg struct {
	Index	int
	Content interface{}
}

type Node struct {
	mu 				sync.Mutex
	clientMu		sync.Mutex
	peers 			[]*rpc.Client
	max_requests	int
	kill_flag		bool
	ecdsaKey		[]byte
	helloSignature	[]byte
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
	waiting 				interface{}
	timeout 				int
	clientBuffer 			string

	applyCh					chan ApplyMsg   // Make sure that the client keeps getting the content out of applyCh
}

type Request struct {
	id int
	seq int
	view int
	reqtype string  //or int?
	msg interface{}  // any kind of msg
	timestamp int64
}


func (nd *Node) execute(am ApplyMsg) {
	// TODO: add msg to applyCh, should be executed in a separate go routine
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

func (nd *Node) executeInOrder() {
	// TODO:
}

func (nd *Node) serverLoop() {
	counter := 0

	for {
		// get events from the buffer

	}
}

func (nd *Node) clean() {
	// TODO: do clean work
}

func (nd *Node) ProcessInit() {

}

func (nd *Node) ProcessCheckpoint() {
	// TODO: Add checkpoint support
}

func (nd *Node) IsInClientLog() {

}

func (nd *Node) AddClientLog() {
	return // We don't have logs for now
}

func (nd *Node) HandleTimeout() {

}

func (nd *Node) NewClientRequest() {

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

	go nd.serverLoop()
	return nd
}