# Guarddog

[![Build Status](https://travis-ci.org/codedokode/guarddog.svg?branch=master)](https://travis-ci.org/codedokode/guarddog)

Guarddog is an utility in Go that executes a program while restricting a set of system calls it is allowed to make. Guarddog is also able to chroot(2) into a given directory and change user and group ids before running the program. This is done so that you don't have to put chroot program into the sandbox.

Guarddog uses seccomp(2) with `SECCOMP_SET_MODE_FILTER` option to put a restriction so the kernel must support it. On attempt to make a call not specified in a list of allowed calls the kernel sends a `SIGKILL` signal that terminates the program.

## Installation

First you need to install Go language compiler. This program is tested with `go1.3.3 linux/386` (can be checked with `go version`).

This program uses [libseccomp](https://github.com/seccomp/libseccomp) so you need to install it beforehands. On Debian or Ubuntu this can be done by running `./scripts/install.sh`. 

Locate the source code so that it is inside your Go workspace in the `src/guarddog` directory. For example, you can create a directory `/tmp/go/src/guarddog/` and copy repository contents there. And then run the following command to set GOPATH:

    export GOPATH=/tmp/go/

By default (for example if you use `go get`) Go will try to install the code into `github.com/codedokode/guarddog` directory and that won't work. Because I don't want to write a github repository URL in every import.

You also have to download a specific version of libseccomp-golang from https://github.com/seccomp/libseccomp-golang/tree/1b506fc7c24eec5a3693cdcbed40d9c226cfc6a1 (you can download newer versions but they are not guaranteered to work). The git repository contents should be copied to `external/github.com/seccomp/libseccomp-golang/` so that the `README` file is located at `external/github.com/seccomp/libseccomp-golang/README`.

You can do this by running these commands from the root of the repository: 

```sh
git clone 'https://github.com/seccomp/libseccomp-golang.git' './external/github.com/seccomp/libseccomp-golang/'
git -C './external/github.com/seccomp/libseccomp-golang/' checkout --detach 1b506fc7c24eec5a3693cdcbed40d9c226cfc6a1
```

## Building

To build a program you can run `./script/build.sh`. This will create a binary named `guarddog`.

## Testing

The program contains unit tests. To test a build run `./scripts/run-tests.sh`

## Usage

Run compiled program with `-help` flag to get help:

    ./guarddog -help

By default no system calls are allowed. You should at least allow `execve` system call or guarddog will be unable to execute a program. You can see the system calls the program is making with `strace` command. 

You can see usage example in file [./scripts/test-sandbox.sh](./scripts/test-sandbox.sh).

Current options are: 

```
Usage: ./guarddog [options] -- command [args]
Options:
  -allow=[]: names of system calls to allow, may be used several times
  -allow-any-syscalls=false: do not apply seccomp syscall filter
  -allow-root=false: allow program to run as root (by default it would refuse to do it)
  -chroot-path="": chroot to a directory before executing program
  -config-file="": read options from this config file. File contains lines like 'some-opti
on = some-value'
  -dump-syscalls=false: print available syscalls names and numbers for current system
  -set-gid=0: switch to this GID
  -set-uid=0: switch to this UID
  -status-fd=0: file descriptor for logging debug and error messsages, default is stderr (
2)
  -trap=false: when making a syscall that is not allowed, send SIGSYS to a program instead
 of SIGKILL. Might be useful for debugging
  -verbose=false: print debugging information
```

## License

The program is distributed under [GNU Affero General Public License](https://www.gnu.org/licenses/agpl-3.0.en.html) version 3.0 or higher. The code in the `external/github.com/seccomp` folder is not a part of this program and is under its own license.

