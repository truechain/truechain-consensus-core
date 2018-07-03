# PBFT-Core
===

*These codes have not gone through reviews. Please use them with cautions*

This code base is an ongoing implementation of Practical Byzantine Fault Tolerance protocol. This PBFT will be the BFT layer in our hybrid consensus design. For this testnet, we will be using PBFT alone to support services and meanwhile working on development and research on the rest of hybrid consensus.

Although there exists a bunch of existing PBFT implementations, we decide to write our own version to get fully control of every details and to conveniently make necessary extensions so that it could fit into the hybrid consensus protocol (which requires more than a standard PBFT).



### Build

#### Step 1

Install [Docker](https://docs.docker.com/install/) and [HyperMake](http://evo-cloud.github.io/hmake/quickguide/install/).


#### Step 2

check hmake command options
```
git clone https://github.com/hixichen/truechain-consensus-core.git
cd truechain-consensus-core
git checkout devel
hmake --targets
```

#### Step 3

```
hmake build
```

The first time, it would download:

- TrueChain's docker image `go-toolchain` from https://hub.docker.com/r/truechain/go-toolchain/
- Dependencies managerment as per `src/vendor/manifest`, which again, could be generated using `gvt fetch` 
   please note that if you need to add package, you could run `gvt fetch` first, then merge the file into `src/vendor/manifest` 

Then, you will get bin/{linux/darwin}/truechain-engine

### Installation && Run

Populate the `/etc/hosts` file with repetitive 5-6 lines containing loopback IP address `127.0.0.1`. 

The entry should look like:
        "127.0.0.1  localhost"

```
$ ./bin/{linux/darwin}/truechain-engine
```

This triggers both server and client subroutines. Also displays progress of key signing, data exchange and ledger log.

```
[.]Loading IP configs...
[ ]%!(EXTRA string=127.0.0.1)[ ]%!(EXTRA string=127.0.0.1)[ ]%!(EXTRA string=127.0.0.1)[ ]%!(EXTRA string=127.0.0.1)[ ]%!(EXTRA string=127.0.0.1)[ ]%!(EXTRA string=127.0.0.1)Get IPList [127.0.0.1 127.0.0.1 127.0.0.1 127.0.0.1 127.0.0.1 127.0.0.1], Ports [40540 40541 40542 40543 40544 40545]
[.]Generated 6 keypairs in /home/arcolife/go/src/github.com/truechain/truechain-consensus-core/keys folder..
127.0.0.1 40540 0
[!]Going to tolerate 1 adversaries
[!]Initial Config &{{5 /home/arcolife/go/src/github.com/truechain/truechain-consensus-core/keys  [127.0.0.1 127.0.0.1 127.0.0.1 127.0.0.1 127.0.0.1 127.0.0.1] [40540 40541 40542 40543 40544 40545] /home/arcolife/hosts 100 6} {0 0} {0 0} [] 40540 100 false <nil> <nil> <nil> <nil> 0 0 5 0 true 1 0 0 0 0 0 0 [] 100 0 0 map[] 600  map[] map[] map[] map[] map[] map[] <nil> <nil> {{0 0} map[]} map[] 0xc420022c00}
fetching file:  sign0.pub
[.]Fetched private keyfetching file:  sign1.pub
fetching file:  sign2.pub
fetching file:  sign3.pub
...
<snip>
```

### CI
To be added

### Deployment

To be added.

### Benchmark

To be added.

### How to contribute

We need contributions from you. You are welcome to create github issues and contribute to the codebase. Developer Guide could be found in `docs/DEV.md`. 

We have a list of important tasks from Foundation. We welcome people who have related background to join our consensus design and implementation.


### LICENSE

The Apache License (2.0).

A copy of the header is included within all files in the codebase along with the full LICENSE txt file in project's root folder.