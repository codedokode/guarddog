package util 

import (
    "errors"
    "fmt"
    "os"
    "guarddog/config"
    /* Though it is deprecated, unix module does not provide exec call */
    "syscall"
    "unsafe"
)

func ChrootInto(path string) error {
    err := syscall.Chroot(path)
    if err != nil {
        return err
    }

    return os.Chdir("/")
}

func ChangeUids(uid int, gid int, allowRoot bool) error {
    // uid == 0 can be set only if allowed
    if uid == 0 && !allowRoot {
        return errors.New("to set uid 0 you must set allow-root option")
    }

    if gid != config.USE_DEFAULT_ID {
        err := syscall.Setresgid(gid, gid, gid)
        if err != nil {
            return fmt.Errorf("failed to set effective, real and saved gid to '%d': %s", gid, err)
        }

        err = syscall.Setgroups([]int{})
        if err != nil {
            return fmt.Errorf("when specifying set-gid options, all supplementary groups are removed. Failed to remove supplementary groups")
        }
    }

    if uid != config.USE_DEFAULT_ID {
        err := syscall.Setresuid(uid, uid, uid)
        if err != nil {
            return fmt.Errorf("failed to set effective, real and saved uid '%d': %s", uid, err)
        }
    }

    if os.Geteuid() == 0 && !allowRoot {
        return errors.New("trying to run as effective uid 0 without allow-root set")
    }

    if os.Getuid() == 0 && !allowRoot {
        return errors.New("trying to run as real uid 0 without allow-root set")
    }

    return nil
}

func MarkAsCloseOnExec(fd int) {
    syscall.CloseOnExec(fd)
}

func SetParentDeathSignal() error {
    _, _, errno := syscall.RawSyscall(
        syscall.SYS_PRCTL, 
        uintptr(syscall.PR_SET_PDEATHSIG), 
        uintptr(syscall.SIGKILL),
        0)

    if errno != 0 {
        return syscall.Errno(errno)
    }
    return nil
    // return unix.Prctl(unix.PR_SET_PDEATHSIG, unix.SIGKILL)
}

func ExecuteProgram(command []string) error {
    var path string = command[0]
    return Exec(path, command, os.Environ())
}

/* Copied from https://github.com/golang/go/blob/8d1d9292ff024f6c7586d27edd2c84c1ca8d9bf5/src/syscall/exec_unix.go#L245 */
// Exec invokes the execve(2) system call.
func Exec(argv0 string, argv []string, envv []string) (err error) {
    argv0p, err := syscall.BytePtrFromString(argv0)
    if err != nil {
        return err
    }
    argvp, err := SlicePtrFromStrings(argv)
    if err != nil {
        return err
    }
    envvp, err := SlicePtrFromStrings(envv)
    if err != nil {
        return err
    }
    _, _, err1 := syscall.RawSyscall(syscall.SYS_EXECVE,
        uintptr(unsafe.Pointer(argv0p)),
        uintptr(unsafe.Pointer(&argvp[0])),
        uintptr(unsafe.Pointer(&envvp[0])))

    if err1 != 0 {
        return syscall.Errno(err1)
    }

    return nil
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


