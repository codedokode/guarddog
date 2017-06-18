#!/bin/bash

# Install necessary libraries under Debian or Ubuntu
set -e 
cd "`dirname "$0"`"
sudo apt-get install libseccomp-dev

