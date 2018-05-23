# PBFT-Core
===

*These codes have not gone through reviews. Please use them with cautions*

This code base is an ongoing implementation of Practical Byzantine Fault Tolerance protocol. This PBFT will be the BFT layer in our hybrid consensus design. For this testnet, we will be using PBFT alone to support services and meanwhile working on development and research on the rest of hybrid consensus.

Although there exists a bunch of existing PBFT implementations, we decide to write our own version to get fully control of every details and to conveniently make necessary extensions so that it could fit into the hybrid consensus protocol (which requires more than a standard PBFT).


### Build & Installation

This project uses [HyperMake](https://github.com/evo-cloud/hmake) to manage dependencies in a containerized environment.

But you could choose to build without using containers. For a very basic sanity test, run the following:

```
./support/scripts/build.sh linux
./bin/linux/truechain-engine
```

Note: you could also use `darwin` as an argument to build.sh instead of `linux` to get a different platform binary. Support for this will be extended soon.

This triggers both server and client subroutines. Also displays progress of key signing, data exchange and ledger log.

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
