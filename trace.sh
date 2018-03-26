#!/bin/bash

go build engine.go
strace -f bash -c './engine' -o trace.log 2>> trace.log 
