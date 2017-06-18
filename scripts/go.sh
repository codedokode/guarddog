#!/bin/bash

# Sets up GOPATH and executes go command
set -e
dir="`dirname "$0"`"
export GOPATH="`realpath "$dir/../../../"`"
exec go "$@"