package config

import (
    "errors"
    "fmt"
    "os"
)

/* 
    Represents program options 

    The options labelled with `option: ... ` can be used 
    as CLI arguments or config file options.

    SomeOption is specified as `some-option` in config or 
    `--some-option` in CLI args.

    `option` tag sets description, options without it will not be parsed
    `multiple` tag allows multiple values
*/
const USE_DEFAULT_ID = -1

type GuarddogOptions struct {
    ConfigFile  string      `cliOnly:"yes" option:"read options from this config file. File contains lines like 'some-option = some-value'"`
    DumpSyscalls bool       `cliOnly:"yes" option:"print available syscalls names and numbers for current system"`
    Verbose     bool        `option:"print debugging information"`

    ChrootPath  string      `option:"chroot to a directory before executing program"`
    Allow       []string    `option:"names of system calls to allow, may be used several times" multiple:"yes"`
    // AllowFromFile []string  `option:"names of files to read syscall list" multiple:"yes"`
    AllowAnySyscalls bool   `option:"do not apply seccomp syscall filter"`
    SetUid      int64       `option:"switch to this UID"`
    SetGid      int64       `option:"switch to this GID"`
    AllowRoot   bool        `option:"allow program to run as root (by default it would refuse to do it)"`
    // Timeout     int64       `option:"timeout in seconds"`

    StatusFd    int64       `option:"file descriptor for logging debug and error messsages, default is stderr (2)"`
    Trap        bool        `option:"when making a syscall that is not allowed, send SIGSYS to a program instead of SIGKILL. Might be useful for debugging"`

    Command     []string    /* tail of option list */
}

func NewGuarddogOptions() *GuarddogOptions {
    opt := new(GuarddogOptions)
    opt.StatusFd = 2
    opt.SetUid = USE_DEFAULT_ID
    opt.SetGid = USE_DEFAULT_ID

    return opt
}

func (opt *GuarddogOptions) Validate() error {

    /* don't check further */
    if opt.DumpSyscalls {
        return nil
    }

    if opt.SetUid == 0 && !opt.AllowRoot {
        return errors.New("to run program with uid = 0 you need to set --allow-root option")
    }

    /* ChrootPath must be valid directory if given */
    if opt.ChrootPath != "" {
        // isDir, reason := doesDirExist(opt.ChrootPath)
        // if !isDir {
        //     return errors.New(
        //         fmt.Sprintf(
        //             "invalid chroot directory %s: %s", 
        //             opt.ChrootPath, 
        //             reason))
        // }
    }   

    // if opt.Timeout < 0 {
    //     return errors.New("timeout cannot be negative")
    // }

    if opt.SetUid < 0 && opt.SetUid != USE_DEFAULT_ID {
        return errors.New("set-uid must be positive")
    }

    if opt.SetGid < 0 && opt.SetGid != USE_DEFAULT_ID {
        return errors.New("set-gid must be positive")
    }

    if opt.Trap && opt.AllowAnySyscalls {
        return errors.New("using -trap along with -allow-any-syscalls makes no sence")
    }

    return nil 
}

func (opt *GuarddogOptions) IsSyscallAllowed (name string) bool {
    return opt.AllowAnySyscalls || containsString(opt.Allow, name)
}

func doesDirExist(path string) (doesExist bool, reason string) {
    stat, err := os.Stat(path)

    if os.IsNotExist(err) {
        return false, "path does not exist"
    }

    if (err != nil) {
        return false, fmt.Sprintf("cannot stat given path: %s", err)
    }    

    if ! stat.IsDir() {
        return false, "path is not a directory"
    }

    return true, ""
}

func containsString(haystack []string, needle string) bool {
    for _, value := range haystack {
        if value == needle {
            return true
        }
    }
    return false
}
