#!/bin/bash
# set -x -e

# pre-prepare
export GOOS=$1

dir=$(dirname $0)
# let sts=0

export TRUE_CONF_DIR='/etc/truechain/' # contains hosts, tunables and logistics configurables
export TRUE_LOG_DIR='/var/log/truechain' # contains all things ledger and logs
export TRUE_LIB_DIR='/var/lib/truechain' # contains keys and all things db

# prepare
if [ ! -d $TRUE_LIB_DIR ]; then
    mkdir -p ${TRUE_LIB_DIR%/}/keys
    chown -R $USER:$USER ${TRUE_LIB_DIR%/}
    # let sts=sts+$?
fi

if [ ! -d $TRUE_LOG_DIR ]; then
    mkdir -p ${TRUE_LOG_DIR%/}/keys
    chown -R $USER:$USER $TRUE_LOG_DIR
    # let sts=sts+$?
fi

if [ ! -d ${TRUE_CONF_DIR%/}keys ]; then
    mkdir -p ${TRUE_CONF_DIR%/}/keys
    chown -R $USER:$USER $TRUE_CONF_DIR
    # let sts=sts+$?
fi

# install
OUTDIR="bin/$GOOS"
cp config/* ${TRUE_CONF_DIR%/}/
cp $OUTDIR/{truechain-engine,pbft-client} $GOBIN/

# sanity checks for installation

# if [ $sts -gt 0 ]; then
#     echo "Unit tests FAILED"
# fi
# exit $sts
