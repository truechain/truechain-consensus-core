# Developer Guide

```
Usage of ./bin/linux/truechain-engine:
  -numquest int
    	number of requests (default 10)

# build
hmake build-linux

# run once for 100 requests
./bin/linux/truechain-engine --numquest 100
```

## Development environment setup

To be added

## Sample runs while building locally with colored output

```
# clear previous junk, rebuild binary and output to fresh log
rm -rf keys/ logs/* && hmake build-linux && script -q -c './bin/linux/truechain-engine 2>&1' > logs/complete_log

# view output (interpreted as binary by Linux, but it's really a color coded log file, dumped by tricking stdout as terminal session)
cat logs/complete_log.txt  | tr -d '\000' | grep New
```

### Ways to automate while debugging..

```
# run for incremental request counts each time
for i in {1..50}; do script -q -c './bin/linux/truechain-engine --numquest '$i' 2>&1' > logs/complete_log && tail logs/complete_log && cat logs/complete_log  | tr -d '\000' | grep REQUEST; done

# or if you'd like to see outputs from same 
for i in {1..10}; do script -q -c './bin/linux/truechain-engine --numquest 17 2>&1' > logs/complete_log && tail logs/complete_log; done

# or with less
less -R logs/complete_log
```

we use `script` while executing the binary and `tr` while reading from it, so we're able to see colors in log file. `grep` mistakes it for a binary by default.

## Debugging

#### using Delve Go Debugger

- [Install devle](https://github.com/derekparker/delve/blob/master/Documentation/installation/linux/install.md)

```
go get -u github.com/derekparker/delve/cmd/dlv
```

### Example Scenario

We get a go panic for `invalid memory address or nil pointer dereference`. The traceback is as follows:

```
[.]Loading IP configs...
Generated 1000 keys in keys/ folder..
[.][0] Entering server loop.
[.]Firing up peer server...
[.][1] Entering server loop.
[.]Firing up peer server...
[.][2] Entering server loop.
[.]Firing up peer server...
[.][0] Ready to listen.
[.][1] Ready to listen.
[.][2] Ready to listen.
[.][0] Begin to setup RPC connections.
[.][1] Begin to setup RPC connections.
[.][2] Begin to setup RPC connections.
[46 255 129 3 1 1 10 80 114 105 118 97 116 101 75 101 121 1 255 130 0 1 2 1 9 80 117 98 108 105 99 75 101 121 1 255 132 0 1 1 68 1 255 134 0 0 0 47 255 131 3 1 1 9 80 117 98 108 105 99 75 101 121 1 255 132 0 1 3 1 5 67 117 114 118 101 1 16 0 1 1 88 1 255 134 0 1 1 89 1 255 134 0 0 0 10 255 133 5 1 2 255 136 0 0 0]
{{<nil> <nil> <nil>} <nil>}
[.]adding signature.
[.]Firing up client executioner...
[.]digest ff2a8a28fad1607e269e6759a564d73eedc19467cb45556a36d45a15d38b4e43df3bf8a71017c68eace5a778ebd0b57b823f0b1ac6a333f379c01e8fe8fda2dd.
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x38 pc=0x5abb41]

goroutine 1 [running]:
crypto/ecdsa.Sign(0x938b00, 0xc420084db0, 0xc42023a030, 0xc4202f8200, 0x80, 0x80, 0x0, 0x0, 0xc420309c48, 0xc420309c48)
	/usr/local/go/src/crypto/ecdsa/ecdsa.go:153 +0x41
_/home/username/truechain-consensus-core/pbft-core.(*Request).addSig(0xc420309d28, 0xc42023a030)
	/home/username/truechain-consensus-core/pbft-core/node.go:277 +0x2ce
_/home/username/truechain-consensus-core/pbft-core.(*Client).NewRequest(0xc42023a000, 0xc4202701c0, 0x9, 0x0)
	/home/username/truechain-consensus-core/pbft-core/client.go:57 +0x1a5
main.main()
	/home/username/truechain-consensus-core/engine.go:55 +0x49b

```

### Steps to debug

#### Start Dlv debugger

```
dlv debug engine.go 
Type 'help' for list of commands.
(dlv) 
```

#### Add breakpoints

```
(dlv) break engine.go:55
Breakpoint 1 set at 0x7e4ed6 for main.main() ./engine.go:55

(dlv) break pbft-core/client.go:57
Breakpoint 2 set at 0x7d27f0 for _/home/username/truechain-consensus-core/pbft-core.(*Client).NewRequest() ./pbft-core/client.go:57

(dlv) break pbft-core/node.go:277
Breakpoint 3 set at 0x7d521e for _/home/username/truechain-consensus-core/pbft-core.(*Request).addSig() ./pbft-core/node.go:277

(dlv) break /usr/local/go/src/crypto/ecdsa/ecdsa.go:153
Breakpoint 4 set at 0x63d0f2 for crypto/ecdsa.Sign() /usr/local/go/src/crypto/ecdsa/ecdsa.go:153
```

#### Start debugging

```
(dlv) continue

[.]Loading IP configs...
Generated 1000 keys in keys/ folder..
[.][0] Entering server loop.
[.]Firing up peer server...
[.][1] Entering server loop.
[.]Firing up peer server...
[.]Firing up peer server...
[.][2] Entering server loop.
[.][0] Ready to listen.
[.][1] Ready to listen.
[.][2] Ready to listen.
[.][0] Begin to setup RPC connections.
[.][1] Begin to setup RPC connections.
[.][2] Begin to setup RPC connections.
[46 255 129 3 1 1 10 80 114 105 118 97 116 101 75 101 121 1 255 130 0 1 2 1 9 80 117 98 108 105 99 75 101 121 1 255 132 0 1 1 68 1 255 134 0 0 0 47 255 131 3 1 1 9 80 117 98 108 105 99 75 101 121 1 255 132 0 1 3 1 5 67 117 114 118 101 1 16 0 1 1 88 1 255 134 0 1 1 89 1 255 134 0 0 0 10 255 133 5 1 2 255 136 0 0 0]
{{<nil> <nil> <nil>} <nil>}
[.]Firing up client executioner...
> main.main() ./engine.go:55 (hits goroutine(1):1 total:1) (PC: 0x7e4ed6)
Warning: debugging optimized function
    50:		}
    51:	
    52:		/////////////////
    53:		cl := pbft.BuildClient(cfg, cfg.Network.IPList[cfg.N], cfg.Network.Ports[cfg.N], 0)
    54:		for k := 0; k < NUM_QUEST; k++ {
=>  55:			cl.NewRequest("Request "+strconv.Itoa(k), int64(k))
    56:		}
    57:		fmt.Println("Finish sending the requests.")
    58:	}

```

### Print some random var


Note:
..can't do func calls but only statements already in call stack, since it's a compiler debugger;
..vis-à-vis Python's interpretted Pdb which allows whatever D-FAQ we wwant.

```
(dlv) print k
0

(dlv) print "Request "+strconv.Itoa(k)
Command failed: no type entry found, use 'types' for a list of valid types
```

### Go next in breakpoint flow

Type `next` or `n`:

```go
(dlv) n
> _/home/username/truechain-consensus-core/pbft-core.(*Client).NewRequest() ./pbft-core/client.go:57 (hits goroutine(1):1 total:1) (PC: 0x7d27f0)
Warning: debugging optimized function
    52:		//broadcast the request
    53:		for i := 0; i < cl.Cfg.N; i++ {
    54:			//req := Request{RequestInner{cl.Cfg.N,0, 0, typeRequest, MsgType(msg), timeStamp, nil}, "", msgSignature{nil, nil}}  // the N-th party is the client
    55:			req := Request{RequestInner{cl.Cfg.N, 0, 0, typeRequest, MsgType(msg), timeStamp}, "", MsgSignature{nil, nil}} // the N-th party is the client
    56:			//req.inner.outer = &req
=>  57:			req.addSig(&cl.privKey)
    58:			arg := ProxyNewClientRequestArg{req, cl.Me}
    59:			reply := ProxyNewClientRequestReply{}
    60:			cl.peers[i].Go("Node.NewClientRequest", arg, &reply, nil)
    61:		}
    62:	}
```

### Print out a struct, let's dive into the issue

```go
(dlv) print cl.privKey
crypto/ecdsa.PrivateKey {
	PublicKey: crypto/ecdsa.PublicKey {
		Curve: crypto/elliptic.Curve nil,
		X: *math/big.Int nil,
		Y: *math/big.Int nil,},
	D: *math/big.Int nil,}

```

hmmm, what's cl anyway?

```go
(dlv) print cl
*_/home/username/truechain-consensus-core/pbft-core.Client {
	IP: "127.0.0.1",
	Port: 40543,
	Index: 0,
	Me: 0,
	Cfg: *_/home/username/truechain-consensus-core/pbft-core.Config {
		N: 3,
		IPList: []string len: 4, cap: 4, [
			"127.0.0.1",
			"127.0.0.1",
			"127.0.0.1",
			"127.0.0.1",
		],
		Ports: []int len: 4, cap: 4, [40540,40541,40542,40543],
		HostsFile: "/home/username/hosts",},
	privKey: crypto/ecdsa.PrivateKey {
		PublicKey: (*crypto/ecdsa.PublicKey)(0xc4201fa030),
		D: *math/big.Int nil,},
	peers: []*net/rpc.Client len: 3, cap: 3, [
		*(*net/rpc.Client)(0xc4201c2300),
		*(*net/rpc.Client)(0xc42026c0c0),
		*(*net/rpc.Client)(0xc4201aa6c0),
	],}
```

hmm, what's `rpc.Client` struct like for `cl.peers` ?

```go
(dlv) print cl.peers
[]*net/rpc.Client len: 3, cap: 3, [
	*{
		codec: net/rpc.ClientCodec(*net/rpc.gobClientCodec) ...,
		reqMutex: (*sync.Mutex)(0xc4201c2310),
		request: (*net/rpc.Request)(0xc4201c2318),
		mutex: (*sync.Mutex)(0xc4201c2338),
		seq: 0,
		pending: map[uint64]*net/rpc.Call [],
		closing: false,
		shutdown: false,},
	*{
		codec: net/rpc.ClientCodec(*net/rpc.gobClientCodec) ...,
		reqMutex: (*sync.Mutex)(0xc42026c0d0),
		request: (*net/rpc.Request)(0xc42026c0d8),
		mutex: (*sync.Mutex)(0xc42026c0f8),
		seq: 0,
		pending: map[uint64]*net/rpc.Call [],
		closing: false,
		shutdown: false,},
	*{
		codec: net/rpc.ClientCodec(*net/rpc.gobClientCodec) ...,
		reqMutex: (*sync.Mutex)(0xc4201aa6d0),
		request: (*net/rpc.Request)(0xc4201aa6d8),
		mutex: (*sync.Mutex)(0xc4201aa6f8),
		seq: 0,
		pending: map[uint64]*net/rpc.Call [],
		closing: false,
		shutdown: false,},
]
```

hmm, nope just 1 is enough:

```go
(dlv) print cl.peers[0]
*net/rpc.Client {
	codec: net/rpc.ClientCodec(*net/rpc.gobClientCodec) *{
		rwc: io.ReadWriteCloser(*net.TCPConn) ...,
		dec: *(*encoding/gob.Decoder)(0xc420158c00),
		enc: *(*encoding/gob.Encoder)(0xc4204f21e0),
		encBuf: *(*bufio.Writer)(0xc4204ef880),},
	reqMutex: sync.Mutex {state: 0, sema: 0},
	request: net/rpc.Request {
		ServiceMethod: "",
		Seq: 0,
		next: *net/rpc.Request nil,},
	mutex: sync.Mutex {state: 0, sema: 0},
	seq: 0,
	pending: map[uint64]*net/rpc.Call [],
	closing: false,
	shutdown: false,}
```

hmm, enough digression (sorry!), again.. what's `cl.privKey` like?

```go
(dlv) print cl.privKey
crypto/ecdsa.PrivateKey {
	PublicKey: crypto/ecdsa.PublicKey {
		Curve: crypto/elliptic.Curve nil,
		X: *math/big.Int nil,
		Y: *math/big.Int nil,},
	D: *math/big.Int nil,}

```

It says this is nil. Meaning, the read method on .dat files for private keys is buggy.

Let's see if we have private keys loaded properly for server ?

```go
(dlv)break engine.go:54
Breakpoint 2 set at 0x7e53a1 for main.main() ./engine.go:54

(dlv) p svList[1].Nd.keyDict
map[int]_/home/username/truechain-consensus-core/pbft-core.keyItem nil

(dlv) p svList[1].Nd.ecdsaKey
*crypto/ecdsa.PrivateKey nil
```

So the private key wasn't loaded for server too. There's some issue with initializeKeys() method in node.go it seems (because that's responsible for reading private keys from the files..

So the following step needs a marshalling from `x509.MarshalECPrivateKey(privateKey)` before we encode this into a `.dat` file.

Per se, the following function needs an additional step for marshalling in node.go:
```go
func (nd *Node) initializeKeys() {
	gob.Register(&ecdsa.PrivateKey{})

	for i := 0; i < nd.N; i++ {
		filename := fmt.Sprintf("sign%v.dat", i)
		b, err := ioutil.ReadFile(path.Join(GetCWD(), "keys/", filename))
		if err != nil {
			MyPrint(3, "Error reading keys %s.\n", filename)
			return
		}
		bufm := bytes.Buffer{}
		bufm.Write(b)
		d := gob.NewDecoder(&bufm)
		sk := ecdsa.PrivateKey{}
		d.Decode(&sk)
		nd.sk = sk
		nd.keyDict[i] = &sk.PublicKey
		if i == nd.id {
			nd.ecdsaKey = &sk
		}
	}
	//nd.ecdsaKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)  // TODO: read from file
```

and so does the following for client.go (from `BuildClient()`):

```go
	cl.peers = peers
	filename := fmt.Sprintf("sign%v.dat", cfg.N)
	kfpath := path.Join(cfg.Logistics.KD, filename)
	MyPrint(2, kfpath + "\n")
	b, err := ioutil.ReadFile(kfpath)
	if err != nil {
		MyPrint(3, "Error reading keys %s.\n", kfpath)
		return nil
	}
	fmt.Println(b)
	bufm := bytes.Buffer{}
	bufm.Write(b)
	gob.Register(&ecdsa.PrivateKey{})
	d := gob.NewDecoder(&bufm)
	sk := ecdsa.PrivateKey{}
	d.Decode(&sk)
	fmt.Println(d)
	// MyPrint(2, sk)
	cl.PrivKey = sk
```

So what we're looking for something like this: https://stackoverflow.com/a/41315404/1332401
