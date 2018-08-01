#!/bin/bash

test_pkgs() {
    for dir in $(find ./trueconsensus \
        -mindepth 1 -maxdepth 1 -type d | grep -vE '/(test|vendor)$') ; do
        echo "$dir/..."
    done
    # for dir in $(find ./trueconsensus/minerva \
    #     -mindepth 1 -maxdepth 1 -type d | grep -vE '/(test|vendor)$') ; do
    #     echo "$dir/..."
    # done
    # for dir in $(find ./trueconsensus/config \
    #     -mindepth 1 -maxdepth 1 -type d | grep -vE '/(test|vendor)$') ; do
    #     echo "$dir/..."
    # done
}

export CONFIGURATION="config/tunables_bft.yaml"

export GOPATH=`pwd`:`pwd`/..
export CGO_ENABLED=1

echo $GOPATH
go test -race $(test_pkgs)
