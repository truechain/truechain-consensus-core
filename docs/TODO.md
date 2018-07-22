### bugs

- [ ] RPC error `panic: open /Users/foo/keys/sign67.pem: too many open files` 
      with high host count (~/hosts) (maybe >40, check limit)
- [ ] issue PR#3 `sync.runtime_SemacquireMutex(0xc4203cc67c, 0x0)` error in lock mechanism 

### overall

- [ ] gRPC secure with TLS
- [ ] add / segregate server/client from engine.go
- [ ] read from config.yaml instead
- [ ] refactor and review, all naive errors like unexportable names in packages, security issues etc..

### transportation.go

- [ ]  implement send/receive messages with RPC

### engine.go > node.go

- [ ] fix this, while doing `go build engine.go`:

```
pbft-core/node.go:251:20: cannot use nil as type DigType in field value
pbft-core/node.go:251:25: cannot use nil as type msgSignature in field value
pbft-core/node.go:643:43: cannot use nil as type Request in field value
pbft-core/node.go:653:43: cannot use nil as type Request in field value
```

### node.go

- [ ] enhance transportation.go
- [ ] read from file, nd.InitializeKeys()
- [ ] change all the int to int64 in case of overflow
- [ ] add msg to applyCh, should be executed in a separate go routine, and we probably want to keep a log for this
- [ ] add timer to try client
- [ ] check if we missed anything and do cleanup
- [ ] Add checkpoint support
- [ ] add counter whenever we found a PREP
- [ ] nd.broadcast(viewChange) // TODO  broadcast view change RPC path.
- [ ] func (nd *Node) NewClientRequest(req Request, clientID int) {  // TODO  change to single arg and single reply
- [ ] nd.broadcast(m)  // TODO  broadcast pre-prepare RPC path.
- [ ] initialize ECDSA keys and hello signature
- [ ] if ok && val.dig == dig {   // TODO  check the diff!
- [ ] if ok && val.dig == dig {   // TODO  check the diff!
- [ ] if req.outer != "" {  // TODO. solve this
- [ ] // TODO  check client message signatures
- [ ] //client_req  = nil  // TODO
- [ ] m  = nd.createRequest(TYPE_PREP, req.inner.seq, MsgType(req.dig))  // TODO  check content!
- [ ] if nd.CheckPrepareMargin(req.dig, req.inner.seq) {  // TODO  check dig vs inner.msg
- [ ] m  = nd.createRequest(TYPE_COMM, req.inner.seq, req.inner.msg) // TODO  check content
- [ ] nd.IncCommDict(m.dig) //TODO  check content
- [ ] if nd.CheckPrepareMargin(req.dig, req.inner.seq) {  // TODO  check dig vs inner.msg
- [ ] m  = nd.createRequest(TYPE_COMM, req.inner.seq, req.inner.msg) // TODO  check content
- [ ] nd.IncCommDict(m.dig) //TODO  check content
- [ ] #TODO (NO PIGGYBACK)
- [ ] // TODO  set up ECDSA

## Testing

WIP:

..add hooks for code coverage, smoke tests, SVT integrations, CI & build status.

## Tests for ...

### PBFT core consensus

- [ ] Unit
- [ ] Functional
- [ ] Integration

### BFT Committee

### Hybrid Consensus

- [ ] Smoke
- [ ] Integration

### SnailChain

- [ ] Unit
- [ ] Functional
