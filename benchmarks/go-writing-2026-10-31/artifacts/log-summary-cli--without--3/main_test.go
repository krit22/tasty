package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLog(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestSummary(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "a.jsonl", "{\"level\":\"WARN\",\"ts\":\"\"}\n{\"level\":\"INFO\",\"ts\":\"anything\",\"other\":42}\n\n")
	writeLog(t, dir, "b.jsonl", "{\"level\":\"ERROR\",\"ts\":\"x\"}\n{\"level\":\"INFO\",\"ts\":\"x\"}")
	writeLog(t, dir, "ignore.txt", "invalid\n")
	if err := os.Mkdir(filepath.Join(dir, "directory.jsonl"), 0700); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{dir}, &out, &diagnostics); code != 0 || diagnostics.Len() != 0 {
		t.Fatalf("code=%d stderr=%s", code, &diagnostics)
	}
	if want := "INFO 2\nERROR 1\nWARN 1\nTOTAL 4\nSKIPPED 1\n"; out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}
}

func TestInvalidRecords(t *testing.T) {
	lines := []string{"", "  ", "bad", "[]", "null", "42", `"text"`, `{}`, `{"level":"","ts":"x"}`, `{"level":3,"ts":"x"}`, `{"level":null,"ts":"x"}`, `{"level":"ok"}`, `{"level":"ok","ts":null}`, `{"level":"ok","ts":4}`, `{"Level":"ok","ts":"x"}`, `{"level":"ok","ts":"x"} {}`}
	s := summary{counts: make(map[string]int)}
	if err := s.read(strings.NewReader(strings.Join(lines, "\n") + "\n")); err != nil {
		t.Fatal(err)
	}
	if s.total != 0 || s.skipped != len(lines) {
		t.Fatalf("total=%d skipped=%d, want 0 and %d", s.total, s.skipped, len(lines))
	}
}

func TestOneMiBRecord(t *testing.T) {
	base := `{"level":"INFO","ts":"","padding":""}`
	line := strings.Replace(base, `"padding":""`, `"padding":"`+strings.Repeat("x", (1<<20)-len(base))+`"`, 1)
	s := summary{counts: make(map[string]int)}
	if err := s.read(strings.NewReader(line + "\n" + line)); err != nil {
		t.Fatal(err)
	}
	if s.total != 2 || s.skipped != 0 {
		t.Fatalf("total=%d skipped=%d", s.total, s.skipped)
	}
}

func TestEmptyDirectory(t *testing.T) {
	var out, diagnostics bytes.Buffer
	if code := run([]string{t.TempDir()}, &out, &diagnostics); code != 0 || out.String() != "TOTAL 0\nSKIPPED 0\n" || diagnostics.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, &out, &diagnostics)
	}
}

func TestArgumentErrors(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	writeLog(t, dir, "file", "")
	for _, args := range [][]string{nil, {dir, dir}, {filepath.Join(dir, "missing")}, {file}} {
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code == 0 || diagnostics.Len() == 0 {
			t.Fatalf("args=%v code=%d stderr=%q", args, code, &diagnostics)
		}
	}
}

func TestUnreadableEntryContinues(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "a-broken.jsonl")
	// A dangling symlink reliably fails to open even when tests run as root.
	if err := os.Symlink(filepath.Join(dir, "missing"), bad); err != nil {
		t.Fatal(err)
	}
	writeLog(t, dir, "z-good.jsonl", `{"level":"INFO","ts":"x"}`)
	var out, diagnostics bytes.Buffer
	if code := run([]string{dir}, &out, &diagnostics); code == 0 || !strings.Contains(diagnostics.String(), bad) || out.String() != "INFO 1\nTOTAL 1\nSKIPPED 0\n" {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, &out, &diagnostics)
	}
}
