package config

import (
    "fmt"
    "reflect"
    "unicode"
)

type Option struct {
    OptionName  string
    FieldName   string
    Kind        reflect.Kind
    IsMultiple  bool
    Desc        string
    IsCliOnly     bool
}

type optionMap map[string]*Option

type OptionList struct {
    options     optionMap
}

func NewOptionList() *OptionList {
    list := new(OptionList)
    list.options = make(optionMap)

    return list
}

func (list *OptionList) Lookup(name string) *Option {
    return list.options[name]
}

func (list *OptionList) Contains(name string) bool {
    _, exists := list.options[name]
    return exists
}

func (list *OptionList) VisitAll(visitor func(name string, opt *Option)) {
    for name, option := range list.options {
        visitor(name, option)
    }
}

func (list *OptionList) Add(option *Option) {    
    name := option.OptionName
    if list.Contains(name) {
        panic(fmt.Sprintf("List already contains option %s", name))
    }

    list.options[name] = option
}

func (list *OptionList) AddFromStruct(ref reflect.Type) {

    for i := 0; i < ref.NumField(); i++ {
        field := ref.Field(i)
        name := field.Name
        optionName := dashifyName(name)
        desc := field.Tag.Get("option")
        isMultiple := field.Tag.Get("multiple") != ""
        isCliOnly := field.Tag.Get("cliOnly") != ""
        kind := field.Type.Kind()        

        if desc == "" {
            continue
        }

        if isMultiple {
            if kind != reflect.Slice {
                panic(fmt.Sprintf("Field %s must be a slice if declared as multiple", name))
            }

            // get internal type, e.g. int instead of []int
            kind = field.Type.Elem().Kind()
        } else {
            if kind == reflect.Slice {
                panic(fmt.Sprintf("Field %s is a slice and must be declared as multiple", name))
            }
        }

        option := Option{
            OptionName: optionName,
            FieldName: name,
            Desc: desc,
            IsMultiple: isMultiple,
            Kind: kind,
            IsCliOnly: isCliOnly,
        }

        list.Add(&option)
    }
}

/* SomeOption -> some-option */
func dashifyName(in string) string {
    runes := []rune(in)
    length := len(runes)

    var out []rune
    for i := 0; i < length; i++ {
        if i > 0 && unicode.IsUpper(runes[i]) && 
                ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
            out = append(out, '-')
        }
        out = append(out, unicode.ToLower(runes[i]))
    }

    return string(out)
}
