package config

import (
    "bufio"
    "fmt"
    "io/ioutil"
    "os"
    "sort"
    "testing"
)

const TMP_DIR = "/tmp"

func TestParsingArgs(t *testing.T) {
    p := createParser()
    args := []string{
        "--verbose", 
        "--allow=open", 
        "--allow=close", 
        "--set-uid=1",
    }
    opt, err := p.ParseNoValidate(args)
    if err != nil {
        t.Fatalf("error while parsing: %s", err)
    }

    if !opt.Verbose {
        t.Fatalf("expected Verbose to be true")
    }

    assertListEqual(t, []string{"open", "close"}, opt.Allow)
    if opt.SetUid != 1 {
        t.Fatalf("expected SetUid to be set")
    }
}

func TestParsingCommand(t *testing.T) {
    p := createParser()
    args := []string{
        "--verbose", 
        "command1", 
        "command2",
    }
    opt, err := p.ParseNoValidate(args)
    if err != nil {
        t.Fatalf("error while parsing: %s", err)
    }

    if !opt.Verbose {
        t.Fatalf("expected Verbose to be true")
    }

    assertListEqual(t, []string{"command1", "command2"}, opt.Command)
}


func TestParsingArgsAndConfig(t *testing.T) {
    name := createTmpFile(`
        allow=read
        allow=write
        set-uid=1
    `)

    defer removeTmpFile(name)
    args := []string{
        "--config-file=" + name,
        "--allow=open",
        "--set-uid=2",
    }

    p := createParser()
    opt, err := p.ParseNoValidate(args)

    if err != nil {
        t.Fatalf("error while parsing: %s", err)
    }

    assertListEqual(t, []string{"open", "write", "read"}, opt.Allow)
    if opt.SetUid != 2 {
        t.Fatalf("expected SetUid to equal %d, %d given", 2, opt.SetUid)
    }
}

func TestParsingWithErrors(t *testing.T) {
    args := []string{
        "--invalid-name=1",
    }

    p := createParser()
    _, err := p.ParseNoValidate(args)

    if err == nil {
        t.Fatalf("expected to get error from invalid argument")
    }
}

func TestCliOnlyOptions(t *testing.T) {
    name := createTmpFile(`
        config-file=config.txt
    `)

    defer removeTmpFile(name)
    args := []string {
        "--config-file=" + name,
    }

    p := createParser()
    _, err := p.ParseNoValidate(args)

    if err == nil {
        t.Fatalf("expected to get error for CLI-only argument")
    }
}

func assertListEqual(t *testing.T, a, b []string) {
    sort.Strings(a)
    sort.Strings(b)

    if len(a) != len(b) {
        t.Fatalf("expected to have array %v, got %v (length different)", a, b)
    }

    for index, value := range a {
        if (b[index] != value) {
            t.Fatalf("expected to have array %v, got %v (values different at index %d)", a, b, index)
        }
    }
}

func createTmpFile(content string) string {
    f, err := ioutil.TempFile(TMP_DIR, "guarddog-test")
    if err != nil {
        panic(fmt.Sprintf("Failed to create tmp file: %s", err))
    }

    defer f.Close()

    writer := bufio.NewWriter(f)
    _, err = writer.WriteString(content)
    writer.Flush()

    if err != nil {
        panic(fmt.Sprintf("Failed to write tmp file: %s", err))
    }

    return f.Name()
}

func removeTmpFile(name string) {
    err := os.Remove(name)
    if err != nil {
        panic(fmt.Sprintf("Failed to delete tmp file '%s': %s", name, err))
    }
}

func createParser() *configuration {
    p := NewConfigurationParser()
    p.testDisableUsage()
    return p
}