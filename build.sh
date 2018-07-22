#!/bin/bash

export GOARCH=amd64
export GOOS=$(go env GOOS)

export CGO_ENABLED=1
export GOPATH=${PWD}

OUTDIR="bin/$GOOS"
mkdir -p "$OUTDIR"

git_commit_hash() {
    echo $(git rev-parse --short HEAD)
}

LDFLAGS="-s -w -X common.GitCommitHash=$(git_commit_hash)"

go build -o "$OUTDIR"/pbft-client \
    -ldflags "$LDFLAGS" \
    ./trueconsensus/client/

go build -o "$OUTDIR"/truechain-engine \
    -ldflags "$LDFLAGS" \
    ./trueconsensus/minerva/
