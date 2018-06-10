#!/bin/bash
set -x

export GOOS=$1
export GOARCH=amd64

git_commit_hash() {
    echo $(git rev-parse --short HEAD)
}

export GOPATH=$GOPATH:`pwd`:`pwd`/..

OUTDIR="bin/$GOOS"
mkdir -p "$OUTDIR"

if [ "$GOOS" -eq "linux" ]; then
    export CGO_ENABLED=1
fi

LDFLAGS="-s -w -X common.GitCommitHash=$(git_commit_hash)"

go build -o "$OUTDIR"/truechain-engine \
    -ldflags "$LDFLAGS" \
    ./src/pbft-core/server/

go build -o "$OUTDIR"/truechain-engine-client \
    -ldflags "$LDFLAGS" \
    ./src/pbft-core/client/

