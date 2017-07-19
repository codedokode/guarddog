#!/bin/bash

# Sets up GOPATH and executes go command
set -e
dir="`dirname "$0"`"
if [ -z ${GOPATH+x} ]
then
    export GOPATH="`realpath "$dir/../../../"`"
fi
exec go "$@"