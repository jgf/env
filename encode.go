// Package env contains the Marshal function for shell environment variables.
// It looks for fields marked with the tag `env:"NAME"` and exports thier value as a shell variable NAME.
// Supported types that can be tagged with `env:"NAME"`: primitive types, structs and pointers to those types.
package env

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type encodeState struct {
	bytes.Buffer
	visited map[reflect.Value]bool
}

// Marshal looks for fields marked with the tag `env:"NAME"` and exports thier value as a shell variable NAME.
// Supported types that can be tagged with `env:"NAME"`: primitive types, structs and pointers to those types.
func Marshal(v interface{}) ([]byte, error) {
	e := &encodeState{visited: make(map[reflect.Value]bool)}
	err := e.marshal(v)
	if err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

func (e *encodeState) marshal(v interface{}) (err error) {
	return e.visitValue(reflect.ValueOf(v), "", false)
}

func (e *encodeState) visitValue(v reflect.Value, tag string, omitEmpty bool) error {
	if !v.IsValid() || e.visited[v] {
		return nil
	}

	e.visited[v] = true

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if tag != "" && !(omitEmpty && isEmptyValue(v)) {
			e.Buffer.WriteString(fmt.Sprintf("export %s='%v'\n", tag, v.Int()))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if tag != "" && !(omitEmpty && isEmptyValue(v)) {
			e.Buffer.WriteString(fmt.Sprintf("export %s='%v'\n", tag, v.Uint()))
		}
	case reflect.Float32, reflect.Float64:
		if tag != "" && !(omitEmpty && isEmptyValue(v)) {
			e.Buffer.WriteString(fmt.Sprintf("export %s='%v'\n", tag, v.Float()))
		}
	case reflect.String:
		if tag != "" && !(omitEmpty && isEmptyValue(v)) {
			e.Buffer.WriteString(fmt.Sprintf("export %s='%v'\n", tag, v.String()))
		}
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			return e.visitValue(v.Elem(), tag, false)
		}
	case reflect.Struct:
		for i := 0; i < v.Type().NumField(); i++ {
			field := v.Field(i)
			envTag := v.Type().Field(i).Tag.Get("env")
			name, opts := parseTag(envTag)
			if tag != "" && name == "" {
				name = tag + "_" + v.Type().Field(i).Name
			}
			err := e.visitValue(field, name, opts.Contains("omitempty"))
			if err != nil {
				return fmt.Errorf("visiting %v: %w", v.Type().Field(i).Name, err)
			}
		}
	default:
		if tag != "" && !(omitEmpty && isEmptyValue(v)) {
			return UnsupportedTypeError(v.Type().String())
		}
	}

	return nil
}

// UnsupportedTypeError is used when an unsupported type is marked to be marshalled.
// Currenlty only primitive types and structs (and pointers to them) are supported.
type UnsupportedTypeError string

// String returns a string representation of the unsuppported type error.
func (e UnsupportedTypeError) String() string {
	return "unsupported type: " + string(e)
}

// Error returns a string representation of the unsuppported type error.
func (e UnsupportedTypeError) Error() string {
	return e.String()
}

type tagOptions string

func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
