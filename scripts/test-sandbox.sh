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

# Returns result in global variable '$output' 
function run_command() {
    local expect_zero="$1"
    shift
    local command="$@"
    echo "Command: $command"
    output=`$command`
    local code=$?

    if [ $expect_zero == "zero" ]
    then 
        if [ "$code" -ne 0 ]
        then
            echo "Expected zero exit code, got $code"
            echo 
            echo "Result of running under strace: "
            strace -f $command
            exit 1
        fi

        echo "Exit code of command was 0"
        return 0
    fi 

    if [ "$expect_zero" == "nonzero" ]
    then
        if [ "$code" -eq 0 ]
        then
            echo "Expected non-zero exit code, got zero"
            echo 
            echo "Result of running under strace: "
            strace -f $command
            exit 1
        fi

        echo "Exit code of command was $code"
        return 0
    fi 

    echo "First argument must be 'zero' or 'nonzero', given '$expect_zero'"
    exit 2
}

# Minimal set of syscalls used by dynamic loader and standard library on i686 arch
# Obtained by running strace -f /bin/true 
# execve() is needed to start the program
ALLOWED_CALLS=( execve brk access mmap2 open fstat64 close read set_thread_area mprotect munmap exit_group )

# Syscalls are different on different archs
ARCH=`uname -m`
if [ "$ARCH" == x86_64 ]
then 
    ALLOWED_CALLS+=( fstat mmap arch_prctl )
fi 

# ALLOWED_CALLS=( execve brk access mmap2 open fstat64 close read set_thread_area mprotect munmap exit_group )
BINARY="$1"
FLAGS=""

[ ! -f "$BINARY" ] && { echo "Fail: path to tested binary not given"; exit 1; } 

allowed_options=()
for syscall in "${ALLOWED_CALLS[@]}"
do 
    allowed_options+=("--allow=$syscall")
done 

echo 
echo "Test: execve works when no filtering is set"
command1="$BINARY $FLAGS -verbose -allow-any-syscalls -- /bin/echo yes"
run_command zero $command1
expect_string "yes" "$output"

echo 
echo "Test: whether echo works with write allowed"
command1="$BINARY $FLAGS -verbose -trap ${allowed_options[@]} -allow=write -- /bin/echo yes"
run_command zero $command1
expect_string "yes" "$output"

echo 
echo "Test: whether echo fails if write is not allowed"
command2="$BINARY $FLAGS -trap ${allowed_options[@]} -- /bin/echo no"
run_command nonzero $command2
expect_string "" "$output"

echo 
echo "Functional tests finished OK"
exit 0

