package config

import (
    "fmt"
    "errors"
    "strconv"
)

/* Helper types to allow multiple int and string options parsing */
type intListFlag []int
type stringListFlag []string

func (option *intListFlag) String() string {
    return fmt.Sprintf("%v", *option)
}

func (option *intListFlag) Set(source string) error {
    value, err := strconv.ParseInt(source, 10, 32)
    if err != nil {
        return err
    }

    *option = append(*option, int(value))
    return nil
}

func (option *intListFlag) Get() interface{} {
    return *option
}

func (option *stringListFlag) String() string {
    return fmt.Sprintf("%s", *option)
}

func (option *stringListFlag) Set(source string) error {
    if source == "" {
        return errors.New("empty string value")
    }

    *option = append(*option, source)
    return nil
}

func (option *stringListFlag) Get() interface{} {
    return *option
}

