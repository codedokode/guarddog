package main 

import (
    "flag"
    "fmt"
    "os"
    "guarddog/config"
    "guarddog/util"
    "guarddog/seccomphelper"
    "guarddog/external/github.com/seccomp/libseccomp-golang"
)

func main() {
    p := config.NewConfigurationParser()
    options, err := p.ParseNoValidate(os.Args[1:])

    if err == flag.ErrHelp {
        // Don't need to call p.PrintUsage()
        os.Exit(0)
    } else if err != nil {
        // Don't need to print fmt.Fprintf(os.Stderr, "%s\n", err)
        os.Exit(1)
    }

    err = options.Validate()    

    if err != nil {
        fmt.Fprintf(os.Stderr, "%s: invalid options: %s\n", config.PROGRAM_NAME, err)
        os.Exit(1)
    }

    if options.StatusFd > 2 {
        util.MarkAsCloseOnExec(int(options.StatusFd))
    }

    logger, err := util.NewLogger(
        int(options.StatusFd), 
        options.Verbose, 
        config.PROGRAM_NAME + ": ")

    if err != nil {
        fmt.Fprintf(os.Stderr, "%s: failed to start logging: %s\n", config.PROGRAM_NAME, err)
        os.Exit(1)
    }

    if options.DumpSyscalls {
        dumpSyscalls()
        os.Exit(0)
    }

    if len(options.Command) > 0 {
        err = executeCommand(logger, options, options.Command)
        if err != nil {
            logger.Error("%s", err)
            os.Exit(1)
        }
    } else {
        logger.Error("command not specified");
        p.PrintUsage()
        os.Exit(1)
    }
}

func dumpSyscalls() {
    debugInfo := seccomphelper.GetLibraryInfo()
    fmt.Printf(
        "# arch %s, libseccomp version %s\n",
        debugInfo.Arch, 
        debugInfo.LibseccompVersion)

    list := seccomphelper.GetSyscallNames()
    for syscallInfo := range list {
        fmt.Printf("%4d %s\n", syscallInfo.Number, syscallInfo.Name)
    }
}

func executeCommand(logger *util.Logger, options *config.GuarddogOptions, command []string) error {

    // We need to be able to allocate memory 
    requiredSyscalls := []string{"execve", "brk", "mmap2", "write"}
    for _, name := range requiredSyscalls {
        if !options.IsSyscallAllowed(name) {
            logger.Info("syscall %s is not allowed, program might fail", name)
        }
    }
    // Check that command is specified as an absolute path to an existing file?

    err := seccomphelper.ExecuteWithSeccomp(
        options.Verbose,
        int(options.StatusFd),
        config.PROGRAM_NAME,
        options.AllowAnySyscalls,
        options.Allow,
        options.Trap,
        command)

    return err

    // filter, err := prepareEnvironment(logger, options)
    // if err != nil {
    //     return err
    // }

    // err = executeWithFilter(logger, filter, /* options.Timeout, */ options.Command)
    // return err 
}

func prepareEnvironment(logger *util.Logger, options *config.GuarddogOptions) (*seccomp.ScmpFilter, error) {

    var filter *seccomp.ScmpFilter
    var err error

    if !options.AllowAnySyscalls {
        filter, err = seccomphelper.PrepareSeccompFilter(options.Allow, options.Trap)
        if err != nil {
            return nil, err
        }
        logger.Info("created seccomp filter")
    } else {
        logger.Info("skip creating seccomp filter")
    }

    if options.ChrootPath != "" {
        err := util.ChrootInto(options.ChrootPath)
        if err != nil {
            return nil, err
        }
        logger.Info("chrooted into '%s'", options.ChrootPath)
    }

    err = util.ChangeUids(int(options.SetUid), int(options.SetGid), options.AllowRoot)
    if err != nil {
        return nil, err
    }

    logger.Info("uid=%d, gid=%d", os.Geteuid(), os.Getegid())

    return filter, nil
}

func executeWithFilter(logger *util.Logger, filter *seccomp.ScmpFilter /* , timeout int */, command []string) error {
    var err error 

    // if timeout != 0 {
    //     err = lib.SetTimeout(timeout)
    //     if err != nil {
    //         return err
    //     }
    //     logger.Info("set timeout to %d seconds", timeout)
    // }

    err = util.SetParentDeathSignal()
    if err != nil {
        return fmt.Errorf("failed to set PR_SET_PDEATHSIG signal: %s", err)
    }

    if filter != nil {
        err = seccomphelper.ApplySeccompFilter(filter)
        if err != nil {
            return err
        }
        logger.Info("applied seccomp filter")
    }

    logger.Info("execting program %v", command)
    err = util.ExecuteProgram(command)
    return err 
}
