package prog

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Bind(v interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		panic(fmt.Errorf("requires ptr, actual: %T", v))
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		panic(fmt.Errorf("requires struct, actual: %T", v))
	}

	t := rv.Type()
	n := rv.NumField()
	for i := 0; i < n; i++ {
		field := rv.Field(i)
		fieldType := t.Field(i)

		if fieldType.Anonymous {
			// recursively parse
			Bind(field.Addr().Interface())
			continue
		}

		s := fieldType.Tag.Get("prog")
		if s == "" {
			continue
		}

		list := strings.SplitN(s, " ", 3)
		if len(list) != 3 {
			panic(fmt.Errorf("invalid option: %v, require 3 parts,actual: %d", fieldType.Name, len(list)))
		}
		flagName := list[0]
		defaulVal := list[1]
		help := list[2]

		switch field.Kind() {
		case reflect.String:
			if defaulVal == "''" {
				defaulVal = ""
			}
			flag.StringVar(field.Addr().Interface().(*string), flagName, defaulVal, help)
		case reflect.Bool:
			v, err := strconv.ParseBool(defaulVal)
			if err != nil {
				panic(fmt.Errorf("parsing %s as bool: invalid default value %s", fieldType.Name, defaulVal))
			}
			flag.BoolVar(field.Addr().Interface().(*bool), flagName, v, help)
		default:
			panic(fmt.Errorf("unsupported type: %s %s", fieldType.Name, field.Type()))
		}
	}

}
