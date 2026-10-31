package render_test

import (
	"encoding/json"
	"testing"

	"example.com/render"
)

func TestRender(t *testing.T) {
	for _, tc := range []struct {
		name    string
		items   []string
		verbose bool
		want    string
	}{
		{name: "plain", items: []string{"alpha", "beta"}, want: "alpha, beta"},
		{name: "verbose", items: []string{"alpha", "beta"}, verbose: true, want: "items: alpha, beta"},
		{name: "nil", want: ""},
		{name: "empty verbose", items: []string{}, verbose: true, want: "items: "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := render.Render(tc.items, tc.verbose); got != tc.want {
				t.Errorf("Render(%q, %t) = %q, want %q", tc.items, tc.verbose, got, tc.want)
			}
		})
	}
}

func TestOptionsFieldsAndJSONTags(t *testing.T) {
	options := render.Options{Verbose: true, Limit: 3}
	data, err := json.Marshal(options)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"verbose":true,"limit":3}`; got != want {
		t.Errorf("json.Marshal(%+v) = %s, want %s", options, got, want)
	}
}
