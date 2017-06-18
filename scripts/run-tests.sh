#!/bin/bash

set -e
cd "`dirname $0`/.."

echo "Running Go unit tests"
./scripts/go.sh test "$@" ./config ./seccomphelper ./util

echo "Building"
./scripts/build.sh -v

echo "Running functinal tests"
./scripts/test-sandbox.sh ./guarddog
code=$?
echo "Exited with code $code"
exit $code