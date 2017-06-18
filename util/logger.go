package util 

import (
    "fmt"
    "os"
)

type Logger struct {
    stream      *os.File
    Verbose     bool
    Prefix      string
}

func NewLogger(fd int, verbose bool, prefix string) (*Logger, error) {
    l := new(Logger)
    l.Verbose = verbose
    l.Prefix = prefix
    if fd > 0 {
        l.stream = os.NewFile(uintptr(fd), "logger-stream")
    }

    return l, nil
}

func (l *Logger) Error(format string, args ...interface{}) {
    if l.stream != nil {
        _, err := fmt.Fprintf(l.stream, l.Prefix + format + "\n", args...) 
        if err != nil {
            panic(err)
        }
    }
}

func (l* Logger) Info(format string, args ...interface{}) {
    if l.stream != nil && l.Verbose {
        _, err := fmt.Fprintf(l.stream, l.Prefix + format + "\n", args...) 
        if err != nil {
            panic(err)
        }
    }
}
