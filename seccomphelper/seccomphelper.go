package seccomphelper

import (
    "guarddog/vendor/github.com/seccomp/libseccomp-golang" 
    "fmt"
    "os"
)

type SeccompInfo struct {
    Arch string
    LibseccompVersion string
}

/* Information about a system call */
type SyscallInfo struct {
    Number int
    Name string
}

func GetLibraryInfo() *SeccompInfo {
    result := new(SeccompInfo)
    archNative, err := seccomp.GetNativeArch()
    if err != nil {
        panic(fmt.Sprintf("GetNativeArch failed: %s", err))
    }
    result.Arch = archNative.String()
    minor, major, micro := seccomp.GetLibraryVersion()
    result.LibseccompVersion = fmt.Sprintf("%d.%d.%d", minor, major, micro)
    return result
}

/*
    Libseccomp doesn't provide a way to iterate through
    a list of system calls so we just loop through
    numbers from 1 to N 

    We use channel as a return value to use the function
    as a generator.
 */
func GetSyscallNames() <-chan SyscallInfo {
    ch := make(chan SyscallInfo)
    go func () {
        for i := 1; i < 1000; i++ {
            syscall := seccomp.ScmpSyscall(i)
            syscallName, error := syscall.GetName()

            if error != nil {
                continue
            }

            ch <- SyscallInfo{i, syscallName}
        }
        close(ch)
    } ()

    return ch
}

/*
    Creates a libseccomp filter that allows only given
    syscalls. The whiteList is a list of syscall names.
 */
func PrepareSeccompFilter(whiteList []string, useTrap bool) (*seccomp.ScmpFilter, error) {
    var actionOnBreakingPolicy = seccomp.ActKill
    if useTrap {
        actionOnBreakingPolicy = seccomp.ActTrap
    }

    filter, err := seccomp.NewFilter(actionOnBreakingPolicy)
    if err != nil {
        return nil, err
    }

    err = filter.SetNoNewPrivsBit(true)
    if err != nil {
        return nil, err
    }

    err = filter.SetBadArchAction(seccomp.ActKill)
    if err != nil {
        return nil, err
    }

    for _, syscallName := range whiteList {
        syscallId, err := seccomp.GetSyscallFromName(syscallName)
        if err != nil {
            return nil, fmt.Errorf("Failed to find a number for syscall name '%s': %s", 
                syscallName, err)
        }
        filter.AddRule(syscallId, seccomp.ActAllow)
    }

    return filter, nil
}

func DebugDumpFilter(filter *seccomp.ScmpFilter) {
    err := filter.ExportPFC(os.Stderr)
    if err != nil {
        panic("Failed to export filter contents")
    }
}

func ApplySeccompFilter(filter *seccomp.ScmpFilter) error {
    return filter.Load()
}
