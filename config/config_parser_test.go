package config

import (
    "fmt"
    "strings"
    "testing"
)

func TestConfigParser(t *testing.T) {
    // Empty config
    testParsing(t, "", map[string]string{})
    testParsing(t, " \n  \n ", map[string]string{})

    // Comments 
    testParsing(t, "# comment", map[string]string{})

    // One-line config
    testParsing(t, "a=1", map[string]string{"a": "1"})

    // Quoted strings and trim
    testParsing(t, "  a  =  ' 1 '  ", map[string]string{"a": " 1 "})

    // Complicated sample
    testParsing(t, `
        # Comment
        a=1

        b='2'
        c="3"

    `, map[string]string{"a": "1", "b": "2", "c": "3"})
}

func TestConfigWithErrors(t *testing.T) {
    // No value
    testReturnsError(t, "test")

    // No key
    testReturnsError(t, " = 1")
}

func testReturnsError(t *testing.T, config string) {
    reader := strings.NewReader(config)
    collector, _ := createCollector()
    err := ParseConfig(reader, collector)

    if err == nil {
        t.Errorf("Expected to get error, got nothing")
    }
}

func createCollector() (func (string, string, int) error, map[string]string) {
    target := map[string]string{}
    collect := func(k string, v string, n int) error {
        target[k] = v
        return nil
    }
    return collect, target
}

func testParsing(t *testing.T, config string, expect map[string]string) {
    reader := strings.NewReader(config)
    collector, result := createCollector()
    err := ParseConfig(reader, collector)

    if err != nil {
        t.Errorf("Parsing config returned error %s", err)
        return 
    }

    diff := compareMaps(expect, result)
    if diff != nil {
        t.Errorf("Expected to get map %v, got %v (%s)", expect, result, diff)
    }
}

func compareMaps(expect map[string]string, result map[string]string) error {
    for k, v := range expect {
        value, exists := result[k]

        if ! exists {
            return fmt.Errorf("missing key '%s'", k)
        }

        if (value != v) {
            return fmt.Errorf("key '%s' has value '%s', expected '%s'", k, value, v)
        }
    }

    for k, _ := range result {
        _, exists := expect[k]
        if ! exists {
            return fmt.Errorf("extra key '%s'", k)
        }
    }

    return nil
}