# truechain-consensus-core
TrueChain Consensus Protocol

[![Waffle.io - Columns and their card count](https://badge.waffle.io/truechain/truechain-consensus-core.svg?columns=all)](https://waffle.io/truechain/truechain-consensus-core)

[![Travis](https://travis-ci.com/truechain/truechain-consensus-core.svg?branch=master)](https://travis-ci.com/truechain/truechain-consensus-core)

# Building the source

## Build

### Step 1

Install [Docker](https://docs.docker.com/install/) and [HyperMake](http://evo-cloud.github.io/hmake/quickguide/install/).

Make sure you have hmake in your `$GOBIN` path.

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


### Step 2

```
git clone https://github.com/truechain/truechain-consensus-core.git
cd truechain-consensus-core
git checkout devel
```

OR

```
go get -u github.com/truechain/truechain-consensus-core
```


### Step 3

Make sure you have `$GOBIN` in `$PATH`:

```
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

And before running, make sure you have the following taken care of, for orchestration:

1. `TRUE_TUNABLES_CONF` corresponds to `/etc/truechain/tunables_bft.yaml`, default's under project's `config/` folder.

2. `TRUE_GENERAL_CONF` corresponds to `/etc/truechain/logistics_bft.cfg`, default's under this project's `config/`

3. `TRUE_NETWORK_CONF` corresponds to `/etc/truechain/hosts`, Default's under this project's `config/`
  This file is populated with repetitive 5-6 lines containing loopback IP address `127.0.0.1`.
  
4. `TRUE_SIMULATION` is as follows:
  - if set to 0 (default) - should tell the project to pickup testbed configurations. 
  - if set to 1 - staging, meaning all CI/CD tests are run before draft run. (dummy functionality at the moment)
  - if set to 1 - production. Will try to connect to boot nodes. (dummy functionality at the moment)

### Run

Server:
```
./bin/{linux/darwin}/truechain-engine
```

Client:
```
$ ./bin/{linux/darwin}/pbft-client -h

Usage of pbft-client:
  -numquest int
    	number of requests (default 10)
```


Optional - To install:

```
./support/scripts/install.sh linux
# then from 1 shell, run
$ truechain-engine
# and a different shell, run
$ pbft-client -numquest 40
```

This triggers both server and client subroutines. Also displays progress of key signing, data exchange and ledger log.

```
2018/07/31 18:19:23 Loaded logistics configuration.
[.]Loading IP configs...

2018/07/31 18:19:23 ---> using following configurations for project:
tunables:
  testbed:
    total: 5
    client_id: 5
    server_id_init: 4

127.0.0.1 49500 0
[!]Going to tolerate 1 adversaries

[!]Initial Node Config &{cfg:0xc4200a6600 mu:{state:0 sema:0} clientMu:{state:0 sema:0} peers:[] port:49500 killFlag:false ListenReady:<nil> SetupReady:<nil> EcdsaKey:<nil> helloSignature:<nil> connections:0 ID:0 N:4 view:0 viewInUse:true f:1 lowBound:0 highBound:0 Primary:0 seq:0 lastExecuted:0 lastStableCheckpoint:0 checkpointProof:[] checkpointInterval:100 vmin:0 vmax:0 waiting:map[] timeout:600 clientBuffer: active:map[] prepared:map[] prepDict:map[] commDict:map[] viewDict:map[] KeyDict:map[] outputLog:<nil> commitLog:<nil> nodeMessageLog:{mu:{state:0 sema:0} content:map[]} clientMessageLog:map[] committedBlock:<nil> txPool:<nil> genesis:<nil> tc:<nil>}

[ ]Genesis block generated: 56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421
...
<snip>
...
[!][3] ProxyProcessCommit 4
[!][1] Committed 4
[!][2] Committed 4
[!][3] Committed 4
```

### CI

```
hmake build test check
```

TODO: add travis yaml

#### managing a missing dependency 

In case you have a new dependency that's not listed in `trueconsensus/vendor/manifest` folder, just run this from `trueconsensus/`:

```
gvt fetch github.com/fatih/color
```

This would add a folder `trueconsensus/vendor` if not already present, and would also generate/append to `trueconsensus/vendor/manifest`.

Note that `gvt fetch <package_name>` updates the file `src/vendor/manifest`.

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
