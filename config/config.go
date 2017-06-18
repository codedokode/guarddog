package config

import (
    "errors"
    "fmt"
    "flag"
    "os"
    "reflect"
    "strconv"
)

const PROGRAM_NAME = "guarddog"

type configuration struct {
    flagSet *flag.FlagSet
    optionList *OptionList
}

func NewConfigurationParser() *configuration {
    cfg := new(configuration)

    cfg.optionList = NewOptionList()
    structType := reflect.TypeOf((*GuarddogOptions)(nil)).Elem()
    cfg.optionList.AddFromStruct(structType)

    cfg.flagSet = createFlagSet(cfg.optionList)
    cfg.flagSet.Usage = func () {
        cfg.PrintUsage()
    }

    return cfg
}

func (cfg *configuration) testDisableUsage() {
    cfg.flagSet.Usage = func () {}
}

func (cfg *configuration) PrintUsage() {
    const preface=`
Guarddog is an utility that executes a program while restricting a set of system calls it is allowed to make. Guarddog is also able to chroot(2) into a given directory and change user and group ids before running the program.

Guarddog uses seccomp(2) with SECCOMP_SET_MODE_FILTER option to put a restriction so the kernel must support it. On attempt to make a call not specified in a list of allowed calls the kernel sends a SIGKILL signal that terminates the program.

Guarddog doesn't search for an executable in PATH. You should specify an absolute path to a program.
`
    fmt.Fprintf(os.Stderr, "%s\n\n", preface)
    fmt.Fprintf(os.Stderr, "Usage: %s [options] -- command [args]\n", os.Args[0])
    fmt.Fprintf(os.Stderr, "Options:\n")
    cfg.flagSet.PrintDefaults()
}

func (cfg *configuration) ParseNoValidate(args []string) (*GuarddogOptions, error) {
    opt := NewGuarddogOptions()
    err := cfg.flagSet.Parse(args)
    if err != nil {
        return nil, fmt.Errorf("invalid CLI arguments: %s", err)
    }

    /* If a config is given, read it */
    if containsFlag(cfg.flagSet, "config-file") {
        configName := cfg.flagSet.Lookup("config-file").Value.String()

        err = cfg.ParseConfigFile(opt, configName)
        if err != nil {
            return nil, 
                fmt.Errorf("error in config file '%s': %s", configName, err)
        }

        opt.ConfigFile = configName
    }

    cfg.updateOptionsFromFlagSet(opt, cfg.flagSet)
    return opt, nil
}

func (cfg *configuration) Parse(args []string) (*GuarddogOptions, error) {

    opt, err := cfg.ParseNoValidate(args)

    if err != nil {
        return nil, err
    }

    err = opt.Validate()    

    if err != nil {
        return nil, fmt.Errorf("invalid options: %s", err)
    }

    return opt, nil
}

func (cfg* configuration) ParseConfigFile(opt *GuarddogOptions, fileName string) error {

    if fileName == "" {
        return errors.New("config file name cannot be empty")
    }

    file, err := os.Open(fileName)
    if err != nil {
        return err
    }
    defer file.Close()

    err = ParseConfig(file, func (key string, value string, num int) error {                
        if ! cfg.optionList.Contains(key) {
            return fmt.Errorf("invalid config option '%s'", key)
        }

        option := cfg.optionList.Lookup(key)

        if option.IsCliOnly {
            return fmt.Errorf("option %s can be used only in CLI args", key)
        }

        err := updateOptionFromString(opt, option, value)
        if err != nil {
            return fmt.Errorf("cannot parse option '%s': %s", key, err)
        }

        return nil
    })

    return err
}

func createFlagSet(options *OptionList) *flag.FlagSet {
    flagSet := flag.NewFlagSet("flags", flag.ContinueOnError)
    options.VisitAll(func (key string, option *Option) {
        addOption(flagSet, option)
    })
    return flagSet
}

func addOption(flagSet *flag.FlagSet, option *Option) {

    optionName := option.OptionName
    desc := option.Desc
    isMultiple := option.IsMultiple

    switch option.Kind {

    case reflect.Bool:
        if (isMultiple) {
            panic(fmt.Sprintf("multibool not implemented, option %s", optionName))
        } else {
            flagSet.Bool(optionName, false, desc)
        }

    case reflect.Float64:
        if (isMultiple) {
            panic(fmt.Sprintf("multifloat not implemented, option %s", optionName))
        } else {
            flagSet.Float64(optionName, 0, desc)
        }

    case reflect.Int64:
        if (isMultiple) {
            list := new(intListFlag)
            flagSet.Var(list, optionName, desc)
        } else {
            flagSet.Int64(optionName, 0, desc)
        }

    case reflect.String:
        if (isMultiple) {
            list := new(stringListFlag)
            flagSet.Var(list, optionName, desc)
        } else {
            flagSet.String(optionName, "", desc)
        }

    default:
        panic(fmt.Sprintf("Unknown kind of options %s", option.Kind))
    }
}

func (cfg *configuration) updateOptionsFromFlagSet(opt *GuarddogOptions, flagSet *flag.FlagSet) {

    options := cfg.optionList

    flagSet.Visit(func (f *flag.Flag) {
        
        if ! options.Contains(f.Name) {
            return
        }

        // Find and update  matching struct field
        option := options.Lookup(f.Name)

        if option.IsMultiple {
            updateMultipleOptionField(opt, option, f.Value)
        } else {
            updateOptionField(opt, option, f.Value)
        }
    })

    // Save tail as command
    opt.Command = flagSet.Args()
}

func updateOptionField(opt *GuarddogOptions, option *Option, v flag.Value) {
    value := reflect.ValueOf(v.(flag.Getter).Get())
    target := getFieldForOption(opt, option)
    target.Set(value)
}

func getFieldForOption(opt *GuarddogOptions, option *Option) reflect.Value {
    structValue := reflect.ValueOf(opt).Elem()
    fieldValue := structValue.FieldByName(option.FieldName)

    return fieldValue
}

func updateMultipleOptionField(opt *GuarddogOptions, option *Option, v flag.Value) {
    target := getFieldForOption(opt, option)
    multiValue := v.(flag.Getter).Get()
    listValue := reflect.ValueOf(multiValue)
    newValue := reflect.AppendSlice(target, listValue)
    target.Set(newValue)
}

func updateOptionFromString(opt *GuarddogOptions, option *Option, value string) error {

    target := getFieldForOption(opt, option)
    parsedValue, err := parseStringValue(option.Kind, value)
    if err != nil {
        return err
    }

    if option.IsMultiple {
        newTarget := reflect.Append(target, parsedValue)
        target.Set(newTarget)
    } else {
        target.Set(parsedValue)
    }

    return nil
}

func parseStringValue(kind reflect.Kind, value string) (reflect.Value, error) {
    switch kind {
        case reflect.Int64:
            i, err := strconv.ParseInt(value, 10, 64)
            return reflect.ValueOf(i), err
        case reflect.Float64:
            f, err := strconv.ParseFloat(value, 64)
            return reflect.ValueOf(f), err
        case reflect.Bool: 
            b, err := strconv.ParseBool(value)
            return reflect.ValueOf(b), err
        case reflect.String:
            return reflect.ValueOf(value), nil
    }

    panic(fmt.Sprintf("Unknown kind: '%s'", kind))
}

func containsFlag(flagSet *flag.FlagSet, name string) bool {
    
    var contains bool

    flagSet.Visit(func (f *flag.Flag) {
        if f.Name == name {
            contains = true
        }
    })

    return contains
}
