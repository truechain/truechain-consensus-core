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

import pb "trueconsensus/fastchain/proto"

// DailyBFTNode - an interface supposed to expose fastchain / committee consensus to an orchestration mechanism
// See - https://github.com/truechain/truechain-consensus-core/issues/26
type DailyBFTNode interface {
	ProxyProcessPrePrepare(arg ProxyProcessPrePrepareArg, reply *ProxyProcessPrePrepareReply) error
	ProxyProcessPrepare(arg ProxyProcessPrepareArg, reply *ProxyProcessPrepareReply) error
	ProxyProcessCommit(arg ProxyProcessCommitArg, reply *ProxyProcessCommitReply) error
	ProxyProcessViewChange(arg ProxyProcessViewChangeArg, reply *ProxyProcessViewChangeReply) error
	ProxyProcessNewView(arg ProxyProcessNewViewArg, reply *ProxyProcessNewViewReply) error
	ProxyProcessCheckpoint(arg ProxyProcessCheckpointArg, reply *ProxyProcessCheckpointReply) error

	processRequest()
	processPrePrepare(req Request, clientID int)
	processPrepare(req Request, clientID int)
	processCommit(req Request)
	processViewChange(req Request, from int)
	processNewView(req Request, clientID int)
	processCheckpoint(req Request, clientID int)

	broadcast(req Request)
	broadcastByRPC(rpcPath string, arg interface{}, reply *[]interface{})

	checkPrepareMargin(dig DigType, seq int) bool
	checkCommittedMargin(dig DigType, req Request) bool

	recordCommit(content string)
	record(content string)
	recordPBFTCommit(req Request)
	recordPBFT(req Request)

	viewProcessCheckPoint(vchecklist *[]Request, lastCheckPoint int) bool
	viewProcessPrepare(vPrepDict viewDict, vPreDict map[int]Request, lastCheckPoint int) (bool, int)
	newViewProcessPrePrepare(prprList []Request) bool
	newViewProcessView(vchangeList []Request) bool

	NewClientRequest(req Request, clientID int)
	createRequest(reqType int, seq int, msg MsgType, block *pb.PbftBlock) Request

	clean()
	initializeKeys()
	suicide()
	serverLoop()
	handleTimeout(dig DigType, view int)
	beforeShutdown()
	setupConnections()

	execute(am ApplyMsg)
	executeInOrder(req Request)

	resetMsgDicts()
	verifyMsg(req Request) bool
	incPrepDict(dig DigType)
	incCommDict(dig DigType)

	isInClientLog(req Request) bool
	isInNodeLog() bool
	addClientLog(req Request)
	addNodeHistory(req Request)
}
