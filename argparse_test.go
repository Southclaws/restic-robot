package main

import (
	"reflect"
	"testing"
)

func TestParseArg(t *testing.T) {
	var tests = map[string][]string{
		"foo":                            {"foo"},
		"foo bar":                        {"foo", "bar"},
		" foo bar ":                      {"foo", "bar"},
		"\"foo bar\" baz":                {"foo bar", "baz"},
		"\\\"foo bar\\\"":                {"\"foo", "bar\""},
		"\\\"foo bar\\\" \"foobar baz\"": {"\"foo", "bar\"", "foobar baz"},
	}
	for input, test := range tests {
		res := parseArg(input)
		if !reflect.DeepEqual(res, test) {
			t.Errorf("result for input '%s' was %+v expected %+v", input, res, test)
		}
	}
}
