package env_test

import (
	"errors"
	"testing"

	"github.com/jgf/env"
)

type sA struct {
	A string `env:"MYVAR"`
	C string
	D int `env:",omitempty"`
}

type sB struct {
	B int `env:"B"`
}

type ssB struct {
	S sB
}

func TestSimpleStruct(t *testing.T) {
	simple := struct {
		A       string `env:"MYVAR"`
		B       int    `env:"B"`
		Ignored int
	}{
		A:       "hallo",
		B:       4711,
		Ignored: 42,
	}

	data, err := env.Marshal(simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := "export MYVAR='hallo'\nexport B='4711'\n"
	if string(data) != expected {
		t.Errorf("marshalled data does not match expectation:\n%v\n%v", string(data), expected)
	}
}

func TestSimpleStructOmitEmpty(t *testing.T) {
	simple := struct {
		A       string `env:"MYVAR,omitempty"`
		B       int    `env:"B,omitempty"`
		Ignored int
	}{
		Ignored: 42,
	}

	data, err := env.Marshal(simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := ""
	if string(data) != expected {
		t.Errorf("marshalled data does not match expectation:\n%v\n%v", string(data), expected)
	}
}

func TestSimplePointerStruct(t *testing.T) {
	simple := struct {
		A       *string `env:"MYVAR"`
		B       *int    `env:"B"`
		Ignored int
	}{
		A:       new(string),
		B:       new(int),
		Ignored: 42,
	}
	*simple.A = "hallo"
	*simple.B = 4711

	data, err := env.Marshal(simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := "export MYVAR='hallo'\nexport B='4711'\n"
	if string(data) != expected {
		t.Errorf("marshalled data does not match expectation:\n%v\n%v", string(data), expected)
	}
}

func TestMultiLayeredStruct(t *testing.T) {
	simple := struct {
		S1      sA  `env:"S1"`
		S2      ssB `env:"S2"`
		Ignored int
	}{
		S1: sA{
			A: "hallo",
			C: "welt",
		},
		S2: ssB{
			S: sB{
				B: 4711,
			},
		},
		Ignored: 42,
	}

	data, err := env.Marshal(simple)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := "export MYVAR='hallo'\nexport S1_C='welt'\nexport B='4711'\n"
	if string(data) != expected {
		t.Errorf("marshalled data does not match expectation:\n%v\n%v", string(data), expected)
	}
}

func TestErrorUnsupportedTypeSlice(t *testing.T) {
	simple := struct {
		A []string `env:"MYSLICE"`
	}{
		A: []string{"hallo", "welt"},
	}

	data, err := env.Marshal(simple)
	if err == nil {
		t.Errorf("expected error did not occur")
	} else if !errors.Is(err, env.UnsupportedTypeError("[]string")) {
		t.Errorf("unexpected error: %v", err)
	}

	if data != nil {
		t.Errorf("expected empty marshalled data, got:\n%v", string(data))
	}
}

func TestErrorUnsupportedTypeArray(t *testing.T) {
	simple := struct {
		A [2]string `env:"MYARR"`
	}{
		A: [2]string{"hallo", "welt"},
	}

	data, err := env.Marshal(simple)
	if err == nil {
		t.Errorf("expected error did not occur")
	} else if !errors.Is(err, env.UnsupportedTypeError("[2]string")) {
		t.Errorf("unexpected error: %v", err)
	}

	if data != nil {
		t.Errorf("expected empty marshalled data, got:\n%v", string(data))
	}
}

func TestErrorUnsupportedTypeMap(t *testing.T) {
	simple := struct {
		A map[string]int `env:"MYMAP"`
	}{
		A: map[string]int{"hallo": 42, "welt": 4711},
	}

	data, err := env.Marshal(simple)
	if err == nil {
		t.Errorf("expected error did not occur")
	} else if !errors.Is(err, env.UnsupportedTypeError("map[string]int")) {
		t.Errorf("unexpected error: %v", err)
	}

	if data != nil {
		t.Errorf("expected empty marshalled data, got:\n%v", string(data))
	}
}

func TestErrorUnsupportedTypeChan(t *testing.T) {
	simple := struct {
		A chan int `env:"MYCHAN"`
	}{
		A: make(chan int),
	}

	data, err := env.Marshal(simple)
	if err == nil {
		t.Errorf("expected error did not occur")
	} else if !errors.Is(err, env.UnsupportedTypeError("chan int")) {
		t.Errorf("unexpected error: %v", err)
	}

	if data != nil {
		t.Errorf("expected empty marshalled data, got:\n%v", string(data))
	}
}
