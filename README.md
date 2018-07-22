# truechain-consensus-core
TrueChain Consensus Protocol

[![Travis](https://travis-ci.com/truechain/truechain-consensus-core.svg?branch=master)](https://travis-ci.com/truechain/truechain-consensus-core)

## Building the source

# PBFT-Core
===

*These codes have not gone through reviews. Please use them with cautions*

This code base is an ongoing implementation of Practical Byzantine Fault Tolerance protocol. This PBFT will be the BFT layer in our hybrid consensus design. For this testnet, we will be using PBFT alone to support services and meanwhile working on development and research on the rest of hybrid consensus.

Although there exists a bunch of existing PBFT implementations, we decide to write our own version to get fully control of every details and to conveniently make necessary extensions so that it could fit into the hybrid consensus protocol (which requires more than a standard PBFT).

### Installation

#### Step 1

Install [Docker](https://docs.docker.com/install/) and [HyperMake](http://evo-cloud.github.io/hmake/quickguide/install/).

Make sure you have hmake in your `$GOBIN` path.

### Step 2

```
go get -u github.com/truechain/truechain-consensus-core
```

### Step 3

Make sure you have `$GOBIN` in `$PATH`:

```
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

Then,

```
cd $GOPATH/src/github.com/truechain/truechain-consensus-core
hmake
cp bin/{linux/darwin}/truechain-engine $GOBIN/
```

### Run

Populate a sample `~/hosts` file with repetitive 5-6 lines containing loopback IP address `127.0.0.1`. 

```
$ truechain-engine
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

### Build

In case you have a new dependency that's not listed in `src/vendor/manifest` folder, just run this from `src/`:

```
gvt fetch github.com/fatih/color
```

This would add a folder `src/vendor` if not already present, and would also generate/append to `src/vendor/manifest`.

#### Building it all with hmake and docker

This project uses:

- [HyperMake](https://github.com/evo-cloud/hmake) to interact with toolchain (containerized environment) and build cross-platform binaries.
- `gvt` to manage dependencies.


```
$ hmake --targets
$ hmake check
$ hmake build
```

The first time, it would download:

- TrueChain's docker image `go-toolchain` from https://hub.docker.com/r/truechain/go-toolchain/
- Dependencies as per `src/vendor/manifest`, which again, could be generated using [gvt](https://github.com/FiloSottile/gvt).

The binaries would be available in `bin/`'s platform-specific folders.

#### Building it all with a shell script

Additionally, you could choose to build without using containers. For a very basic sanity test, run the following:

```
./support/scripts/build.sh {linux/darwin}
```

Note: you could also use `darwin` as an argument to build.sh instead of `linux` to get an OSX binary. Support for more will be extended soon.


### Deployment

To be added.

### Benchmark

To be added.

## Feedback

Feedback is greatly appreciated. We're a blend of `py-go-c++` devs looking to merge ideas and we may err at times in the realm of language paradigms and correct design approach. We're hoping to find you (yes you!) and incentivize you for auditing our codebase, polish the erm out of it and teach us how to fish in this process. Feel free to open issues / contact us on our channels. 

## Contributing

We need contributions from you. You are welcome to create github issues and contribute to the codebase.
We have a list of important tasks from Truechain Foundation. We welcome people who have related background to join
our consensus design and implementation.

The maintainers actively manage the issues list, and try to highlight issues suitable for newcomers.
The project follows the typical GitHub pull request model.
See [CONTRIBUTIONS.md](CONTRIBUTIONS.md) for more details.
Before starting any work, please either comment on an existing issue, or file a new one.

Track our [Milestones here](https://github.com/truechain/truechain-consensus-core/milestones/)
And for different codebases of truechain engineering, [here's an explanation](https://github.com/truechain/truechain-consensus-core/milestone/2)

## Community / Mailing lists

Join our gitter channel for live discussions and clarifications.

1. For contributions/general engineering discussions and all things PR/codebases [truechain-net/engg-foss-global](https://gitter.im/truechain-net/engg-foss-global)
2. For events, announcements and so on [truechain-net/community](https://gitter.im/truechain-net/community)
3. For research, [truechain-net/research](https://gitter.im/truechain-net/research)
4. Architecture at [truechain-net/architecture](https://gitter.im/truechain-net/architecture)

Subscribe to the [google groups mailing list](https://groups.google.com/forum/#!forum/truechain) to receive regular updates and post discussions 

### On Social Media

- Main site: https://www.truechain.pro/en/
  - News from Truechain https://www.truechain.pro/en/news
- Twitter: https://twitter.com/truechaingroup
- Facebook: https://www.facebook.com/TrueChaingroup/
- Telegram: http://www.t.me/truechainglobal/

## LICENSE

The Apache License (2.0).

A copy of the header is included within all files in the codebase along with the full [LICENSE](LICENSE) txt file in project's root folder.
