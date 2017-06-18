#!/bin/bash

# Uploads files to a remote host via SSH
set -e
cd `dirname "$0"`
[ -z "$1" ] && { echo "Usage: ./script user@host"; exit; }
TARGET_DIR=/tmp/go/src/guarddog
rsync --rsync-path="mkdir -p $TARGET_DIR && rsync" -r ../ "$1":$TARGET_DIR/
