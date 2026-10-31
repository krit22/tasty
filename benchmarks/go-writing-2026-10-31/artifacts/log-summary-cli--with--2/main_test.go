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

func TestRunSummary(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "a.jsonl", "{\"level\":\"warn\",\"ts\":\"\"}\n{\"level\":\"info\",\"ts\":\"now\",\"extra\":true}\n\ninvalid\n")
	writeFixture(t, dir, "b.jsonl", "{\"level\":\"error\",\"ts\":\"x\"}\r\n{\"level\":\"info\",\"ts\":\"x\"}")
	writeFixture(t, dir, "ignored.txt", "invalid\n")
	if err := os.Mkdir(filepath.Join(dir, "ignored.jsonl"), 0700); err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	code := run([]string{dir}, &out, &diagnostic)
	want := "info 2\nerror 1\nwarn 1\nTOTAL 4\nSKIPPED 2\n"
	if code != 0 || out.String() != want || diagnostic.Len() != 0 {
		t.Fatalf("run = %d, %q, %q; want 0, %q, empty stderr", code, &out, &diagnostic, want)
	}
}

func TestRecordValidation(t *testing.T) {
	for _, line := range []string{"", " \t\r\n", "null", "[]", "42", "{}", `{"level":"","ts":"x"}`, `{"level":null,"ts":"x"}`, `{"level":1,"ts":"x"}`, `{"level":"info"}`, `{"level":"info","ts":null}`, `{"level":"info","ts":5}`, `{"level":"info","ts":"x"} {}`, `{"Level":"info","ts":"x"}`} {
		t.Run(line, func(t *testing.T) {
			if _, ok := recordLevel([]byte(line)); ok {
				t.Fatalf("accepted %q", line)
			}
		})
	}
}

func TestLargeRecord(t *testing.T) {
	const size = 1 << 20
	prefix := `{"level":"info","ts":"","extra":"`
	line := prefix + strings.Repeat("x", size-len(prefix)-2) + `"}`
	dir := t.TempDir()
	writeFixture(t, dir, "large.jsonl", line+"\n"+line)
	var out, diagnostic bytes.Buffer
	if code := run([]string{dir}, &out, &diagnostic); code != 0 || out.String() != "info 2\nTOTAL 2\nSKIPPED 0\n" {
		t.Fatalf("run = %d, %q, %q", code, &out, &diagnostic)
	}
}

func TestEmptyDirectory(t *testing.T) {
	var out, diagnostic bytes.Buffer
	if code := run([]string{t.TempDir()}, &out, &diagnostic); code != 0 || out.String() != "TOTAL 0\nSKIPPED 0\n" {
		t.Fatalf("run = %d, %q, %q", code, &out, &diagnostic)
	}
}

func TestArgumentErrors(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{nil, {dir, dir}, {filepath.Join(dir, "missing")}} {
		var out, diagnostic bytes.Buffer
		if code := run(args, &out, &diagnostic); code == 0 || diagnostic.Len() == 0 {
			t.Fatalf("run(%q) = %d, stderr %q", args, code, &diagnostic)
		}
	}
}

func TestUnreadableEntryContinues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.jsonl")
	if err := os.Symlink(filepath.Join(dir, "missing"), path); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, dir, "good.jsonl", `{"level":"ok","ts":""}`)
	var out, diagnostic bytes.Buffer
	if code := run([]string{dir}, &out, &diagnostic); code == 0 || !strings.Contains(diagnostic.String(), path) || out.String() != "ok 1\nTOTAL 1\nSKIPPED 0\n" {
		t.Fatalf("run = %d, %q, %q", code, &out, &diagnostic)
	}
}

type failingIO struct{}

func (failingIO) Read([]byte) (int, error)  { return 0, io.ErrUnexpectedEOF }
func (failingIO) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestIOErrors(t *testing.T) {
	if err := readRecords(failingIO{}, make(map[string]uint64), new(uint64), new(uint64)); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("read error = %v", err)
	}
	var diagnostic bytes.Buffer
	if code := run([]string{t.TempDir()}, failingIO{}, &diagnostic); code == 0 || !strings.Contains(diagnostic.String(), "write output") {
		t.Fatalf("run = %d, %q", code, &diagnostic)
	}
}

func writeFixture(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
}
