package config

import (
    "bufio"
    "errors"
    "io"
    "fmt"
    "strings"
)

type ConfigVisitor func(key string, value string, lineNumber int) error

func ParseConfig(reader io.Reader, visitor ConfigVisitor) error {
    bufReader := bufio.NewReader(reader)
    lineNumber := 1

    for {
        line, err := bufReader.ReadString('\n')
        line = strings.TrimSpace(line)

        if !isConfigCommentOrEmpty(line) {
            key, value, parseErr := parseConfigString(line)

            if parseErr != nil {
                return fmt.Errorf("syntax error: %s at line %d", parseErr, lineNumber)
            }

            err := visitor(key, value, lineNumber)
            if err != nil {
                return fmt.Errorf("line %d: %s", lineNumber, err)
            }
        }

        lineNumber++

        // err != nil means EOF or error
        if err == io.EOF {
            return nil
        } else if err != nil {
            return err
        }
    }
}

func isConfigCommentOrEmpty(line string) bool {
    return strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || line == ""
}

func parseConfigString(line string) (key string, value string, err error) {
    parts := strings.SplitN(line, "=", 2)
    if len(parts) < 2 {
        return "", "", errors.New("no equal sign")
    }

    key = strings.TrimSpace(parts[0])
    value = strings.TrimSpace(parts[1])

    if key == "" {
        return "", "", errors.New("key is empty")
    }

    value = removeQuotes(value)

    return key, value, nil
}

func removeQuotes(value string) string {
    if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) && len(value) > 1 {
        value = value[1:len(value) - 1]
    } else if strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`) && len(value) > 1 {
        value = value[1:len(value) - 1]
    }

    return value
}

