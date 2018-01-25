#!/bin/bash

set -e
cd "`dirname $0`/.."

# Strict checks
export GODEBUG='cgocheck=2'

echo "Running go vet"
# Do not stop if something fails
./scripts/go.sh vet ./... || true

echo "Running Go unit tests"
./scripts/go.sh test "$@" ./config ./seccomphelper ./util

echo "Building"
# Disable optimizations for easier debugging
./scripts/build.sh -v -ccflags="-N" -gcflags="-N -l"

echo "Running functional tests"
./scripts/test-sandbox.sh ./guarddog
code=$?
echo "Exited with code $code"
exit $code