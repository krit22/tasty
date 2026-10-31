package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixture(t *testing.T, dir, name, data string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestRunSummary(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "a.jsonl", "{\"level\":\"INFO\",\"ts\":\"\"}\n\ninvalid\n{\"level\":\"ERROR\",\"ts\":\"x\"}\n")
	writeFixture(t, dir, "b.jsonl", "{\"level\":\"WARN\",\"ts\":\"x\"}\r\n{\"level\":\"INFO\",\"ts\":\"x\",\"extra\":true}")
	writeFixture(t, dir, "ignored.txt", "invalid\n")
	if err := os.Mkdir(filepath.Join(dir, "ignored.jsonl"), 0700); err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	code := run([]string{dir}, &out, &diagnostic)
	want := "INFO 2\nERROR 1\nWARN 1\nTOTAL 4\nSKIPPED 2\n"
	if code != 0 || out.String() != want || diagnostic.Len() != 0 {
		t.Fatalf("run = %d, %q, stderr %q; want 0, %q", code, out.String(), diagnostic.String(), want)
	}
}

func TestRecordValidation(t *testing.T) {
	for _, line := range []string{"", " \t\r\n", "null", "[]", "42", "{}", `{"level":"","ts":"x"}`, `{"level":1,"ts":"x"}`, `{"level":null,"ts":"x"}`, `{"level":"A"}`, `{"level":"A","ts":null}`, `{"level":"A","ts":1}`, `{"Level":"A","ts":"x"}`, `{"level":"A","ts":"x"} {}`, `{"level":"A","ts":"x"`} {
		t.Run(line, func(t *testing.T) {
			if level, ok := recordLevel([]byte(line)); ok {
				t.Fatalf("accepted %q with level %q", line, level)
			}
		})
	}
}

func TestLargeRecord(t *testing.T) {
	prefix, suffix := `{"level":"INFO","ts":"","extra":"`, `"}`
	line := prefix + strings.Repeat("x", (1<<20)-len(prefix)-len(suffix)) + suffix
	counts := make(map[string]uint64)
	var skipped uint64
	if err := readRecords(strings.NewReader(line+"\n"+line), counts, &skipped); err != nil {
		t.Fatal(err)
	}
	if counts["INFO"] != 2 || skipped != 0 {
		t.Fatalf("counts = %v, skipped = %d", counts, skipped)
	}
}

func TestRunEmptyAndArguments(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{{dir}, nil, {dir, dir}, {filepath.Join(dir, "missing")}} {
		var out, diagnostic bytes.Buffer
		code := run(args, &out, &diagnostic)
		if len(args) == 1 && args[0] == dir {
			if code != 0 || out.String() != "TOTAL 0\nSKIPPED 0\n" || diagnostic.Len() != 0 {
				t.Fatalf("empty directory: code %d, output %q, stderr %q", code, out.String(), diagnostic.String())
			}
		} else if code == 0 || diagnostic.Len() == 0 {
			t.Fatalf("args %v: code %d, stderr %q", args, code, diagnostic.String())
		}
	}
}

func TestUnreadableEntryContinues(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "a.jsonl")
	// A dangling symlink fails to open even when tests run as root.
	if err := os.Symlink(filepath.Join(dir, "missing"), badPath); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, dir, "z.jsonl", `{"level":"OK","ts":""}`)
	var out, diagnostic bytes.Buffer
	if code := run([]string{dir}, &out, &diagnostic); code == 0 {
		t.Fatal("unreadable entry returned success")
	}
	if !strings.Contains(diagnostic.String(), badPath) || out.String() != "OK 1\nTOTAL 1\nSKIPPED 0\n" {
		t.Fatalf("output %q, stderr %q", out.String(), diagnostic.String())
	}
}

type failingIO struct{}

func (failingIO) Read([]byte) (int, error)  { return 0, errors.New("read failed") }
func (failingIO) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestIOFailures(t *testing.T) {
	if err := readRecords(failingIO{}, make(map[string]uint64), new(uint64)); err == nil {
		t.Fatal("read error was lost")
	}
	if code := run([]string{t.TempDir()}, failingIO{}, io.Discard); code == 0 {
		t.Fatal("write error returned success")
	}
}
