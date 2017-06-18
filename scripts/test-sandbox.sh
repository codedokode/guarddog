# !/bin/bash

# Test whether program can restrict allowed syscalls set
# set -e 

function expect_string() {
    if [ "$1" != "$2" ]
    then 
        echo "Test failed, expected:"
        echo "$1" 
        echo "Got: "
        echo "$2"
        echo 
        exit 1
    fi
}

# Minimal set of syscalls used by dynamic loader and standard library on i686 arch
# Obtained by running strace -f /bin/true 
# execve() is needed to start the program
ALLOWED_CALLS=( execve brk access mmap2 open fstat64 close read set_thread_area mprotect munmap exit_group )
BINARY="$1"
FLAGS=""

[ ! -f "$BINARY" ] && { echo "Fail: path to tested binary not given"; exit 1; } 

allowed_options=()
for syscall in "${ALLOWED_CALLS[@]}"
do 
    allowed_options+=("--allow=$syscall")
done 

echo "Test: execve works when no filtering is set"
command1="$BINARY $FLAGS -verbose -allow-any-syscalls -- /bin/echo yes"
echo "Command: $command1"
output1=`$command1` 
code1=$?
if [ "$code1" -ne 0 ]
then 
    echo "Expected exit code 0, got $code1"
    exit 1
fi
expect_string "yes" "$output1"


# echo with write should be successful
echo "Test: whether echo works with write allowed"
command1="$BINARY $FLAGS -verbose -trap ${allowed_options[@]} --allow=write -- /bin/echo yes"
echo "Command: $command1"
output1=`$command1`
code1=$?
if [ "$code1" -ne 0 ]
then 
    echo "Expected exit code 0, got $code1"
    exit 1
fi
expect_string "yes" "$output1"

# echo should fail
echo "Test: whether echo fails if write is not allowed"
command2="$BINARY $FLAGS -trap ${allowed_options[@]} -- /bin/echo no"
echo "Command: $command2"
output2=`$command2`
code2=$?
if [ "$code1" -ne 0 ]
then 
    echo "Expected non-zero exit code, got $code2"
    exit 1
fi
expect_string "" "$output2"

echo "Tests finished OK"
exit 0
