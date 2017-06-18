package seccomphelper

import (
    "errors"
    "guarddog/vendor/github.com/seccomp/libseccomp-golang" 
    "syscall"
    "unsafe"
)

// Using C here because applying filter can 
// interfere with Go runtime

/*
#cgo pkg-config: libseccomp

#include <stdlib.h>

int executeProgramWithFilter(
        int verbose, 
        int loggerFd, 
        char const *loggerTag,
        int allowAnySyscalls,
        int const allowedCallNumbers[], 
        int useTrap, 
        char const* const argv[],
        char* errorBuffer,
        int errorBufferLength
);

 */
import "C"

func ExecuteWithSeccomp(
    verbose bool, 
    loggerFd int, 
    loggerTag string,
    allowAnySyscalls bool,
    allowedCalls []string,
    useTrap bool,
    command []string) error {

    argv, err := SlicePtrFromStrings(command)
    if err != nil {
        return err
    }

    // Create a slice of system call numbers
    var allowedCallNumbers = make([]C.int, len(allowedCalls) + 1)
    var i = 0
    for _, syscallName := range allowedCalls {
        syscallId, err := seccomp.GetSyscallFromName(syscallName)
        if err != nil {
            return err
        }

        allowedCallNumbers[i] = C.int(syscallId)
        i++
    }

    // Make last item empty
    allowedCallNumbers[i] = 0;

    // Buffer to write an error message
    const ERROR_BUFFER_LEN = 2048
    var errorBufferC = (*C.char)(C.malloc(ERROR_BUFFER_LEN + 1))
    defer C.free(unsafe.Pointer(errorBufferC))

    if errorBufferC == nil {
        return errors.New("Failed to allocate memory for error message")
    }

    var loggerTagC = C.CString(loggerTag)
    defer C.free(unsafe.Pointer(loggerTagC))
    
    _ = C.executeProgramWithFilter(
        C.int(bool2int(verbose)),
        C.int(loggerFd),
        loggerTagC,
        C.int(bool2int(allowAnySyscalls)),
        &allowedCallNumbers[0],
        C.int(bool2int(useTrap)),
        (**C.char)(unsafe.Pointer(&argv[0])),
        errorBufferC,
        C.int(ERROR_BUFFER_LEN))

    var errorText = C.GoString(errorBufferC)
    return errors.New(errorText)
}

func bool2int(b bool) int {
    if b {
        return 1
    } else {
        return 0
    }
}

// SlicePtrFromStrings converts a slice of strings to a slice of
// pointers to NUL-terminated byte arrays. If any string contains
// a NUL byte, it returns (nil, EINVAL).
func SlicePtrFromStrings(ss []string) ([]*byte, error) {
    var err error
    bb := make([]*byte, len(ss)+1)
    for i := 0; i < len(ss); i++ {
        bb[i], err = syscall.BytePtrFromString(ss[i])
        if err != nil {
            return nil, err
        }
    }
    bb[len(ss)] = nil
    return bb, nil
}
