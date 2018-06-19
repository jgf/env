package env

import (
	"runtime"
	"reflect"
	"bytes"
	"fmt"
	"strings"
)

type encodeState struct {
	bytes.Buffer
	visited map[reflect.Value]bool
}

func Marshal(v interface{}) ([]byte, error) {
	e := &encodeState{visited:make(map[reflect.Value]bool)}
	err := e.marshal(v)
	if err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

func (e *encodeState) marshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			err = r.(error)
		}
	}()
	e.reflectValue(reflect.ValueOf(v), "", false)
	return nil
}

func (e *encodeState) reflectValue(v reflect.Value, tag string, omitEmpty bool) {
	if !v.IsValid() || e.visited[v] {
		return
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
			e.reflectValue(v.Elem(), "", false)
		}
	case reflect.Struct:
		for i := 0; i < v.Type().NumField(); i++ {
			field := v.Field(i)
			tag := v.Type().Field(i).Tag.Get("env")
			name, opts := parseTag(tag)
			e.reflectValue(field, name, opts.Contains("omitempty"))
		}
	case reflect.Array:
		if tag != "" && !(omitEmpty && isEmptyValue(v)) {
			panic(UnsupportedTypeError{v.Type()})
		}
	case reflect.Slice:
		if tag != "" && !(omitEmpty && isEmptyValue(v)) {
			panic(UnsupportedTypeError{v.Type()})
		}
	case reflect.Map:
		if tag != "" && !(omitEmpty && isEmptyValue(v)) {
			panic(UnsupportedTypeError{v.Type()})
		}
	default:
		panic(UnsupportedTypeError{v.Type()})
	}
}

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "env: unsupported type: " + e.Type.String()
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
