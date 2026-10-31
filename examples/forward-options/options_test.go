package render_test

import (
	"encoding/json"
	"errors"
	"testing"

	"example.com/render"
)

func TestParseOptions(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
		want render.Options
	}{
		{name: "fields", data: `{"verbose":true,"limit":12}`, want: render.Options{Verbose: true, Limit: 12}},
		{name: "zero limit", data: `{"limit":0}`},
		{name: "absent fields", data: `{}`},
		{name: "null", data: `null`},
		{name: "unknown field", data: `{"extra":true,"limit":2}`, want: render.Options{Limit: 2}},
		{name: "whitespace", data: " \n {\"limit\":4} \t", want: render.Options{Limit: 4}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := render.ParseOptions([]byte(tc.data))
			if err != nil {
				t.Fatalf("ParseOptions(%q) returned error: %v", tc.data, err)
			}
			if got != tc.want {
				t.Errorf("ParseOptions(%q) = %+v, want %+v", tc.data, got, tc.want)
			}
		})
	}
}

func TestParseOptionsRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
	}{
		{name: "negative limit", data: `{"verbose":true,"limit":-1}`},
		{name: "empty input"},
		{name: "malformed JSON", data: `{"limit":`},
		{name: "wrong root", data: `[]`},
		{name: "wrong boolean type", data: `{"verbose":1}`},
		{name: "wrong integer type", data: `{"limit":"2"}`},
		{name: "fractional limit", data: `{"limit":1.5}`},
		{name: "integer overflow", data: `{"limit":9223372036854775808}`},
		{name: "trailing JSON value", data: `{"limit":1} {"limit":2}`},
		{name: "trailing garbage", data: `{"limit":1} x`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := render.ParseOptions([]byte(tc.data))
			if err == nil {
				t.Fatalf("ParseOptions(%q) = %+v, nil; want error", tc.data, got)
			}
			if got != (render.Options{}) {
				t.Errorf("ParseOptions(%q) = %+v with error %v, want zero Options", tc.data, got, err)
			}
		})
	}
}

func TestParseOptionsPreservesSyntaxError(t *testing.T) {
	_, err := render.ParseOptions([]byte(`{"limit":`))
	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Fatalf("ParseOptions malformed JSON error = %v (%T), want *json.SyntaxError cause", err, err)
	}
}

func TestParseOptionsPreservesTypeError(t *testing.T) {
	_, err := render.ParseOptions([]byte(`{"limit":"2"}`))
	var typeError *json.UnmarshalTypeError
	if !errors.As(err, &typeError) {
		t.Fatalf("ParseOptions wrong limit type error = %v (%T), want *json.UnmarshalTypeError cause", err, err)
	}
	if typeError.Field != "limit" {
		t.Errorf("error field = %q, want limit", typeError.Field)
	}
}
