package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func invoke(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(os.Getenv("TASTY_EVAL_BINARY"), args...)
	var out, errOut strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	return out.String(), errOut.String(), err
}

func write(t *testing.T, dir, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestTastyAggregation(t *testing.T) {
	d := t.TempDir()
	write(t, d, "a.jsonl", "{\"level\":\"WARN\",\"ts\":\"a\"}\n{\"level\":\"INFO\",\"ts\":\"b\"}\n")
	write(t, d, "b.jsonl", "{\"level\":\"INFO\",\"ts\":\"c\"}\n{\"level\":\"ERROR\",\"ts\":\"d\"}\n")
	write(t, d, "ignored.txt", "not json")
	out, stderr, err := invoke(t, d)
	want := "INFO 2\nERROR 1\nWARN 1\nTOTAL 4\nSKIPPED 0\n"
	if err != nil || stderr != "" || out != want {
		t.Fatalf("got %q, stderr %q, err %v; want %q", out, stderr, err, want)
	}
}

func TestTastyInvalidRecords(t *testing.T) {
	d := t.TempDir()
	write(t, d, "a.jsonl", "\nnot-json\nnull\n[]\n{}\n{\"level\":3,\"ts\":\"a\"}\n{\"level\":\"\",\"ts\":\"a\"}\n{\"level\":\"INFO\"}\n{\"level\":\"INFO\",\"ts\":\"a\"}\n")
	out, _, err := invoke(t, d)
	if err != nil || out != "INFO 1\nTOTAL 1\nSKIPPED 8\n" {
		t.Fatalf("invalid record policy: output=%q err=%v", out, err)
	}
}

func TestTastyLongRecord(t *testing.T) {
	d := t.TempDir()
	write(t, d, "long.jsonl", "{\"level\":\"INFO\",\"ts\":\"a\",\"extra\":\""+strings.Repeat("x", 100000)+"\"}\n")
	out, _, err := invoke(t, d)
	if err != nil || out != "INFO 1\nTOTAL 1\nSKIPPED 0\n" {
		t.Fatalf("large record: output=%q err=%v", out, err)
	}
}

func TestTastyEmptyDirectory(t *testing.T) {
	out, _, err := invoke(t, t.TempDir())
	if err != nil || out != "TOTAL 0\nSKIPPED 0\n" {
		t.Fatalf("empty directory: %q, %v", out, err)
	}
}

func TestTastyArgumentErrors(t *testing.T) {
	for _, args := range [][]string{nil, {"a", "b"}, {filepath.Join(t.TempDir(), "missing")}} {
		_, stderr, err := invoke(t, args...)
		if err == nil || strings.TrimSpace(stderr) == "" || strings.Contains(stderr, "panic:") {
			t.Errorf("args=%q stderr=%q err=%v", args, stderr, err)
		}
	}
}

func TestTastyUnreadableEntry(t *testing.T) {
	d := t.TempDir()
	write(t, d, "good.jsonl", "{\"level\":\"INFO\",\"ts\":\"a\"}\n")
	bad := filepath.Join(d, "missing.jsonl")
	if err := os.Symlink(filepath.Join(d, "does-not-exist"), bad); err != nil {
		t.Fatal(err)
	}
	out, stderr, err := invoke(t, d)
	if err == nil || !strings.Contains(stderr, bad) || out != "INFO 1\nTOTAL 1\nSKIPPED 0\n" {
		t.Fatalf("partial failure: out=%q stderr=%q err=%v", out, stderr, err)
	}
}
