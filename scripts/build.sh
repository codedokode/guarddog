#!/bin/bash

set -e 
cd "`dirname $0`/.."
exec ./scripts/go.sh build "$@" guarddog.go
