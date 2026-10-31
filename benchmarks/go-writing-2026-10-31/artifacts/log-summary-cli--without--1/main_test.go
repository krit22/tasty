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

func TestRun(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "a.jsonl", "{\"level\":\"INFO\",\"ts\":\"\"}\n{\"level\":\"WARN\",\"ts\":\"x\"}\n\ninvalid\n")
	writeLog(t, dir, "b.jsonl", "{\"level\":\"ERROR\",\"ts\":\"x\"}\r\n{\"level\":\"INFO\",\"ts\":\"x\",\"extra\":1e9999}")
	writeLog(t, dir, "ignored.txt", "invalid\n")
	if err := os.Mkdir(filepath.Join(dir, "ignored.jsonl"), 0700); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	status := run([]string{dir}, &out, &diagnostics)
	want := "INFO 2\nERROR 1\nWARN 1\nTOTAL 4\nSKIPPED 2\n"
	if status != 0 || out.String() != want || diagnostics.Len() != 0 {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, &out, &diagnostics)
	}
}

func TestInvalidRecords(t *testing.T) {
	lines := []string{"", "   ", "null", "[]", "42", `"text"`, "{}", `{"level":"","ts":"x"}`, `{"level":1,"ts":"x"}`, `{"level":null,"ts":"x"}`, `{"level":"INFO"}`, `{"level":"INFO","ts":null}`, `{"level":"INFO","ts":1}`, `{"level":"INFO","ts":"x"} {}`, `{"level":"INFO","ts":"x"`}
	counts := make(map[string]int)
	total, skipped, err := summarize(strings.NewReader(strings.Join(lines, "\n")+"\n"), counts)
	if err != nil || total != 0 || skipped != len(lines) || len(counts) != 0 {
		t.Fatalf("total=%d skipped=%d counts=%v err=%v", total, skipped, counts, err)
	}
}

func TestOneMiBRecord(t *testing.T) {
	base := `{"level":"INFO","ts":"","data":""}`
	record := strings.Replace(base, `"data":""`, `"data":"`+strings.Repeat("x", (1<<20)-len(base))+`"`, 1)
	for _, suffix := range []string{"", "\n", "\r\n"} {
		counts := make(map[string]int)
		total, skipped, err := summarize(strings.NewReader(record+suffix), counts)
		if err != nil || total != 1 || skipped != 0 || counts["INFO"] != 1 {
			t.Fatalf("suffix=%q total=%d skipped=%d err=%v", suffix, total, skipped, err)
		}
	}
}

func TestEmptyDirectory(t *testing.T) {
	var out, diagnostics bytes.Buffer
	if code := run([]string{t.TempDir()}, &out, &diagnostics); code != 0 || out.String() != "TOTAL 0\nSKIPPED 0\n" || diagnostics.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q code=%d", &out, &diagnostics, code)
	}
}

func TestArgumentErrors(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "file", "")
	for _, args := range [][]string{nil, {dir, dir}, {filepath.Join(dir, "missing")}, {filepath.Join(dir, "file")}} {
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code == 0 || diagnostics.Len() == 0 || out.Len() != 0 {
			t.Fatalf("args=%v stdout=%q stderr=%q code=%d", args, &out, &diagnostics, code)
		}
	}
}

func TestUnreadableEntryContinues(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "a.jsonl")
	// A dangling symlink fails to open even when tests run as root.
	if err := os.Symlink(filepath.Join(dir, "missing"), bad); err != nil {
		t.Fatal(err)
	}
	writeLog(t, dir, "z.jsonl", `{"level":"OK","ts":"x"}`)
	var out, diagnostics bytes.Buffer
	if code := run([]string{dir}, &out, &diagnostics); code == 0 || out.String() != "OK 1\nTOTAL 1\nSKIPPED 0\n" || !strings.Contains(diagnostics.String(), bad) {
		t.Fatalf("stdout=%q stderr=%q code=%d", &out, &diagnostics, code)
	}
}
