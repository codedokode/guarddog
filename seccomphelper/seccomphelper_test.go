package seccomphelper

import (
    "testing"
)

func TestGetLibraryInfo(t *testing.T) {
    // Test we can call this function
    _ = GetLibraryInfo()
}

func TestGetSyscalls(t *testing.T) {
    syscalls := []string{}
    for syscallInfo := range GetSyscallNames() {
        syscalls = append(syscalls, syscallInfo.Name)
    }

    if len(syscalls) < 10 {
        t.Fatalf("Too little syscalls: %d", len(syscalls))
    }
}

func TestPrepareFilter(t *testing.T) {
    whitelist := []string{"write", "read", "exit"}
    _, err := PrepareSeccompFilter(whitelist, true)
    if err != nil {
        t.Fatalf("Failed to prepare a filter: %s", err)
    }
}
