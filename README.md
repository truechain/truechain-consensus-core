# PBFT-Core
===

*These codes have not gone through reviews. Please use them with cautions*

This code base is an ongoing implementation of Practical Byzantine Fault Tolerance protocol. This PBFT will be the BFT layer in our hybrid consensus design. For this testnet, we will be using PBFT alone to support services and meanwhile working on development and research on the rest of hybrid consensus.

Although there exists a bunch of existing PBFT implementations, we decide to write our own version to get fully control of every details and to conveniently make necessary extensions so that it could fit into the hybrid consensus protocol (which requires more than a standard PBFT).

### Benchmark

To be added.

### How to re-use

Run the following:

```
go build engine.go
./engine
```

This triggers both server and client subroutines and displays progress of key signing, data exchange and ledger log is written to the same folder.

### Deployment

To be added.

### How to contribute

We need contributions from you. You are welcome to create github issues and contribute to the codebase.

We have a list of important tasks from Foundation. We welcome people who have related background to join our consensus design and implementation.

### LICENSE

The MIT License, a copy of which is included within all files in the codebase.
