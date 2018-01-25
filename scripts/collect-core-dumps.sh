#!/bin/bash

# Finds all core dumps in current directory and prints them
# Call it from project's root directory
set -e

shopt -s nullglob 
for file in core*
do echo "Corefile found: $file, trying to get binary name" 
    BINARIES="dummy-file echo guarddog"
    binary=
    for test_binary in $BINARIES
    do 
        if file "$file" | grep -q "$test_binary" > /dev/null
        then
            binary="$test_binary"
        fi
    done

    echo "Found binary: $binary"
    gdb -c "$file" "$binary" -iex 'set auto-load safe-path /' \
        -ex "thread apply all bt" \
        -ex "set pagination 0" \
        -ex 'printf "\nLast signal info:\n\n"' \
        -ex 'p $_siginfo' \
        -batch 
    echo 
done
