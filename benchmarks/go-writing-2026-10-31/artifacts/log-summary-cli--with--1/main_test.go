package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func putFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestRunSummary(t *testing.T) {
	dir := t.TempDir()
	putFile(t, dir, "a.jsonl", "{\"level\":\"WARN\",\"ts\":\"\"}\n{\"level\":\"INFO\",\"ts\":\"today\",\"extra\":42}\n\ninvalid\n")
	putFile(t, dir, "b.jsonl", "{\"level\":\"ERROR\",\"ts\":\"x\"}\r\n{\"level\":\"INFO\",\"ts\":\"x\"}")
	putFile(t, dir, "ignored.txt", "invalid\n")
	if err := os.Mkdir(filepath.Join(dir, "ignored.jsonl"), 0700); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	code := run([]string{dir}, &out, &diagnostics)
	want := "INFO 2\nERROR 1\nWARN 1\nTOTAL 4\nSKIPPED 2\n"
	if code != 0 || out.String() != want || diagnostics.Len() != 0 {
		t.Fatalf("run = %d, %q, stderr %q; want 0, %q, empty stderr", code, out.String(), diagnostics.String(), want)
	}
}

func TestRecordValidation(t *testing.T) {
	for _, line := range []string{"", " \t\r\n", "null", "[]", "42", "{}", `{"level":"","ts":"x"}`, `{"level":1,"ts":"x"}`, `{"level":null,"ts":"x"}`, `{"level":"INFO"}`, `{"level":"INFO","ts":null}`, `{"level":"INFO","ts":123}`, `{"level":"INFO","ts":"x"} {}`, `{"LEVEL":"INFO","ts":"x"}`} {
		t.Run(line, func(t *testing.T) {
			if level, ok := recordLevel([]byte(line)); ok {
				t.Fatalf("accepted invalid record as %q", level)
			}
		})
	}
}

func TestOneMiBRecord(t *testing.T) {
	prefix, suffix := `{"level":"INFO","ts":"","data":"`, `"}`
	line := prefix + strings.Repeat("x", (1<<20)-len(prefix)-len(suffix)) + suffix
	s := summary{counts: make(map[string]uint64)}
	if err := s.read(strings.NewReader(line + "\n" + line)); err != nil {
		t.Fatal(err)
	}
	if s.total != 2 || s.skipped != 0 {
		t.Fatalf("total=%d skipped=%d; want 2, 0", s.total, s.skipped)
	}
}

func TestEmptyAndInvalidArguments(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{{dir}, nil, {dir, dir}, {filepath.Join(dir, "missing")}} {
		var out, diagnostics bytes.Buffer
		code := run(args, &out, &diagnostics)
		if len(args) == 1 && args[0] == dir {
			if code != 0 || out.String() != "TOTAL 0\nSKIPPED 0\n" || diagnostics.Len() != 0 {
				t.Fatalf("empty directory: code=%d out=%q stderr=%q", code, out.String(), diagnostics.String())
			}
		} else if code == 0 || diagnostics.Len() == 0 {
			t.Fatalf("args %q: code=%d stderr=%q", args, code, diagnostics.String())
		}
	}
}

func TestUnreadableEntryContinues(t *testing.T) {
	dir := t.TempDir()
	broken := filepath.Join(dir, "a.jsonl")
	if err := os.Symlink(filepath.Join(dir, "missing"), broken); err != nil {
		t.Fatal(err)
	}
	putFile(t, dir, "b.jsonl", `{"level":"INFO","ts":""}`)
	var out, diagnostics bytes.Buffer
	if code := run([]string{dir}, &out, &diagnostics); code == 0 {
		t.Fatal("unreadable entry returned success")
	}
	if !strings.Contains(diagnostics.String(), broken) || out.String() != "INFO 1\nTOTAL 1\nSKIPPED 0\n" {
		t.Fatalf("out=%q stderr=%q", out.String(), diagnostics.String())
	}
}

type failingIO struct{}

func (failingIO) Read([]byte) (int, error)  { return 0, errors.New("read failed") }
func (failingIO) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestIOErrors(t *testing.T) {
	s := summary{counts: make(map[string]uint64)}
	if err := s.read(failingIO{}); err == nil {
		t.Fatal("read error lost")
	}
	var diagnostics bytes.Buffer
	if code := run([]string{t.TempDir()}, failingIO{}, &diagnostics); code == 0 || diagnostics.Len() == 0 {
		t.Fatalf("write failure: code=%d stderr=%q", code, diagnostics.String())
	}
}
