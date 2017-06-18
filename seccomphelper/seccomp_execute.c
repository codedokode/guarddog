#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <seccomp.h>
#include <errno.h>
#include <unistd.h>

extern char **environ;

/**
 * Creates and loads a BPF seccomp filter
 * allowing only specific system calls
 *
 * Returns 0 on success, 1 on error
 */
int createAndLoadFilter(
    int const allowedCallNumbers[],
    int useTrap,
    char* errorBuffer,
    int errorBufferLength    
) {
    scmp_filter_ctx filterContext = NULL;
    int result = 0;
    int isError = 0;
    const int *currentCall;

    int actionOnBreakingPolicy = SCMP_ACT_KILL;
    if (useTrap) {
        actionOnBreakingPolicy = SCMP_ACT_TRAP;
    }

    filterContext = seccomp_init(actionOnBreakingPolicy);
    if (!filterContext) {
        snprintf(
            errorBuffer, 
            errorBufferLength, 
            "seccomp_init() failed"
        );
        return 1;
    }

    // Set NO_NEW_PRIVS bit
    result = seccomp_attr_set(filterContext, SCMP_FLTATR_CTL_NNP, 1);
    if (result != 0) {
        snprintf(
            errorBuffer, 
            errorBufferLength, 
            "seccomp_attr_set(SCMP_FLTATR_CTL_NNP) error %d: %s", 
            result, 
            strerror(-result)
        );
        isError = 1;
        goto release;
    }
    
    // Set bad architecture action
    result = seccomp_attr_set(filterContext, SCMP_FLTATR_ACT_BADARCH, SCMP_ACT_KILL);
    if (result != 0) {
        snprintf(
            errorBuffer, 
            errorBufferLength, 
            "seccomp_attr_set(SCMP_FLTATR_ACT_BADARCH) error %d: %s", 
            result, 
            strerror(-result)
        );
        isError = 1;
        goto release;
    }

    // Iterate through a list of allowed system call numbers
    for (currentCall = allowedCallNumbers; *currentCall; currentCall++) {
        result = seccomp_rule_add(filterContext, SCMP_ACT_ALLOW, *currentCall, 0);
        if (result != 0) {
            snprintf(
                errorBuffer,
                errorBufferLength,
                "seccomp_rule_add() failed for call %d with code %d: %s",
                *currentCall,
                result,
                strerror(-result)
            );

            isError = 1;
            goto release;
        }
    }

    result = seccomp_load(filterContext);
    if (result != 0) {
        snprintf(
            errorBuffer,
            errorBufferLength,
            "seccomp_load() failed with code %d: %s",
            result,
            strerror(-result)
        );

        isError = 1;
        goto release;
    }

    
    release: 
    seccomp_release(filterContext);
    return isError;
};

/*
    Creates a libseccomp filter, loads it and executes
    the given program. Written in C to avoid side effects
    of applying seccomp filter on Go runtime.

    This function is not supposed to return if everything 
    is OK.

    Returns 1 on error
 */
int executeProgramWithFilter(
        int verbose, 
        int loggerFd, 
        char const *loggerTag,
        int allowAnySyscalls,
        int const allowedCallNumbers[], 
        int useTrap, 
        char* const argv[],
        char* errorBuffer,
        int errorBufferLength) {

    int result;
    FILE *loggerFile;

    // Clear buffer
    errorBuffer[0] = '\0';

    // We are never going to close this FILE* object so 
    // underlying file descriptor will not be closed too
    loggerFile = fdopen(loggerFd, "a");
    if (!loggerFile) {
        snprintf(
            errorBuffer,
            errorBufferLength,
            "fdopen() for fd %d (logger) failed: %s",
            loggerFd,
            strerror(errno)
        );

        return 0;
    }

    if (!allowAnySyscalls) {
        result = createAndLoadFilter(
            allowedCallNumbers,
            useTrap,
            errorBuffer,
            errorBufferLength
        );

        if (result != 0) {
            return 1;
        }
    }

    if (verbose) {
        char* const *currentArg;
        fprintf(loggerFile, "%s: Applied seccomp policy\n", loggerTag);
        fprintf(loggerFile, "%s: Executing command [", loggerTag);
        for (currentArg = argv; *currentArg; currentArg++) {
            if (currentArg != argv) {
                // Add space except first argument
                fprintf(loggerFile, " ");
            }
            fprintf(loggerFile, "%s", *currentArg);
        }
        fprintf(loggerFile, "]\n");
    }

    // Now call execve
    result = execve(argv[0], argv, environ);

    // We should not get here
    snprintf(
        errorBuffer,
        errorBufferLength,
        "execve() failed with code %d: %s",
        errno,
        strerror(errno)
    );

    fprintf(
        loggerFile, 
        "%s: execve() failed with code %d: %s", 
        loggerTag,
        errno,
        strerror(errno)
    );

    return 1;
};


