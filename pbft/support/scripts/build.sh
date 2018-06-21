#!/bin/bash
set -x

export GOOS=$1
export GOARCH=amd64
export PROJECTPATH="$GOPATH/src/github.com/truechain/truechain-consensus-core/pbft"

cd $PROJECTPATH

git_commit_hash() {
    echo $(git rev-parse --short HEAD)
}

protoc -I src/pbft-core/fastchain/ \
          src/pbft-core/fastchain/fastchain.proto \
          --go_out=plugins=grpc:src/pbft-core/fastchain/

export GOPATH=$GOPATH:`pwd`:`pwd`/..

OUTDIR="bin/$GOOS"
mkdir -p "$OUTDIR"

if [ "$GOOS" -eq "linux" ]; then
    export CGO_ENABLED=1
fi

LDFLAGS="-s -w -X common.GitCommitHash=$(git_commit_hash)"

go build -o "$OUTDIR"/pbft-client \
    -ldflags "$LDFLAGS" \
    ./src/pbft-core/client/

go build -o "$OUTDIR"/truechain-engine \
    -ldflags "$LDFLAGS" \
    ./src/pbft-core/pbft-sim-engine/
