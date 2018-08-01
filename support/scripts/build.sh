#!/bin/bash
set -x

export GOOS=$1
export GOARCH=amd64
export CGO_ENABLED=1

protoc -I trueconsensus/fastchain/proto/ \
          trueconsensus/fastchain/proto/fastchain.proto \
          --go_out=plugins=grpc:trueconsensus/fastchain/proto/

git_commit_hash() {
    echo $(git rev-parse --short HEAD)
}

export GOPATH=$GOPATH:`pwd`:`pwd`/..

OUTDIR="bin/$GOOS"
mkdir -p "$OUTDIR"

if [ "$GOOS" = "darwin" ]; then
    export CC=o64-clang
    export CXX=o64-clang++
fi

if [ "$GOOS" = "windows" ]; then
    export CC=x86_64-w64-mingw32-gcc-posix
    export CXX=x86_64-w64-mingw32-g++-posix
fi

LDFLAGS="-s -w -X common.GitCommitHash=$(git_commit_hash)"

go build -o "$OUTDIR"/pbft-client \
    -ldflags "$LDFLAGS" \
    ./trueconsensus/client/

go build -o "$OUTDIR"/truechain-engine \
    -ldflags "$LDFLAGS" \
    ./trueconsensus/minerva/
