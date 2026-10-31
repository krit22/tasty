package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLog(t *testing.T, dir, name, data string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestSummary(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "a.jsonl", "{\"level\":\"WARN\",\"ts\":\"\"}\r\n{\"level\":\"INFO\",\"ts\":\"now\"}\n\ninvalid\n")
	writeLog(t, dir, "b.jsonl", "{\"level\":\"ERROR\",\"ts\":\"x\",\"extra\":1e999}\n{\"level\":\"INFO\",\"ts\":\"y\"}")
	writeLog(t, dir, "ignore.txt", "invalid\n")
	if err := os.Mkdir(filepath.Join(dir, "nested.jsonl"), 0700); err != nil {
		t.Fatal(err)
	}
	writeLog(t, filepath.Join(dir, "nested.jsonl"), "hidden.jsonl", "invalid\n")
	var out, errs bytes.Buffer
	if code := run([]string{dir}, &out, &errs); code != 0 || errs.Len() != 0 {
		t.Fatalf("code=%d stderr=%q", code, &errs)
	}
	if want := "INFO 2\nERROR 1\nWARN 1\nTOTAL 4\nSKIPPED 2\n"; out.String() != want {
		t.Fatalf("got %q, want %q", &out, want)
	}
}

func TestRecordValidation(t *testing.T) {
	for _, input := range []string{
		"", " \t\r\n", "null", "[]", "42", `"text"`, "{", "{}",
		`{"level":"","ts":"x"}`, `{"level":1,"ts":"x"}`,
		`{"level":null,"ts":"x"}`, `{"level":"INFO"}`,
		`{"level":"INFO","ts":null}`, `{"level":"INFO","ts":5}`,
		`{"level":"INFO","ts":{}}`, `{"level":"INFO","ts":"x"} {}`,
	} {
		t.Run(input, func(t *testing.T) {
			if _, valid := recordLevel([]byte(input)); valid {
				t.Fatalf("accepted %q", input)
			}
		})
	}
	for _, input := range []string{`{"level":"info","ts":""}`, `{"level":" ","ts":"anything","other":[true]}`} {
		if _, valid := recordLevel([]byte(input)); !valid {
			t.Fatalf("rejected %q", input)
		}
	}
}

func TestLargeRecords(t *testing.T) {
	for _, newline := range []string{"", "\n"} {
		t.Run("ending="+newline, func(t *testing.T) {
			dir := t.TempDir()
			prefix, suffix := `{"level":"BIG","ts":"","data":"`, `"}`
			record := prefix + strings.Repeat("x", (1<<20)-len(prefix)-len(suffix)) + suffix
			writeLog(t, dir, "large.jsonl", record+newline)
			var out, errs bytes.Buffer
			if code := run([]string{dir}, &out, &errs); code != 0 || out.String() != "BIG 1\nTOTAL 1\nSKIPPED 0\n" {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, &out, &errs)
			}
		})
	}
}

func TestEmptyDirectory(t *testing.T) {
	var out, errs bytes.Buffer
	if code := run([]string{t.TempDir()}, &out, &errs); code != 0 || out.String() != "TOTAL 0\nSKIPPED 0\n" {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, &out, &errs)
	}
}

func TestBadArguments(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "file", "")
	for _, args := range [][]string{nil, {dir, dir}, {filepath.Join(dir, "missing")}, {filepath.Join(dir, "file")}} {
		var out, errs bytes.Buffer
		if code := run(args, &out, &errs); code == 0 || errs.Len() == 0 || out.Len() != 0 {
			t.Fatalf("args=%q code=%d stdout=%q stderr=%q", args, code, &out, &errs)
		}
	}
}

func TestUnreadableEntryContinues(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "a-broken.jsonl")
	// A dangling symlink fails to open even when tests run as root.
	if err := os.Symlink(filepath.Join(dir, "missing"), badPath); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	writeLog(t, dir, "z-good.jsonl", "{\"level\":\"INFO\",\"ts\":\"x\"}\n")
	var out, errs bytes.Buffer
	if code := run([]string{dir}, &out, &errs); code == 0 || !strings.Contains(errs.String(), badPath) || out.String() != "INFO 1\nTOTAL 1\nSKIPPED 0\n" {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, &out, &errs)
	}
}
