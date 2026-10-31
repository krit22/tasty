package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type summary struct {
	counts  map[string]uint64
	total   uint64
	skipped uint64
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: logsum DIRECTORY")
		return 1
	}
	entries, err := os.ReadDir(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "logsum: %v\n", err)
		return 1
	}
	s := summary{counts: make(map[string]uint64)}
	status := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		path := filepath.Join(args[0], entry.Name())
		if err := s.readFile(path); err != nil {
			fmt.Fprintf(stderr, "logsum: %s: %v\n", path, err)
			status = 1
		}
	}
	if err := s.write(stdout); err != nil {
		fmt.Fprintf(stderr, "logsum: write output: %v\n", err)
		return 1
	}
	return status
}

func (s *summary) readFile(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	return s.read(f)
}

func (s *summary) read(r io.Reader) error {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			level, valid := recordLevel(line)
			if valid {
				s.counts[level]++
				s.total++
			} else {
				s.skipped++
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func recordLevel(line []byte) (string, bool) {
	var fields map[string]json.RawMessage
	if len(bytes.TrimSpace(line)) == 0 || json.Unmarshal(line, &fields) != nil {
		return "", false
	}
	var level, ts *string
	if json.Unmarshal(fields["level"], &level) != nil || level == nil || *level == "" {
		return "", false
	}
	if json.Unmarshal(fields["ts"], &ts) != nil || ts == nil {
		return "", false
	}
	return *level, true
}

func (s *summary) write(w io.Writer) error {
	levels := make([]string, 0, len(s.counts))
	for level := range s.counts {
		levels = append(levels, level)
	}
	sort.Slice(levels, func(i, j int) bool {
		if s.counts[levels[i]] != s.counts[levels[j]] {
			return s.counts[levels[i]] > s.counts[levels[j]]
		}
		return levels[i] < levels[j]
	})
	for _, level := range levels {
		if _, err := fmt.Fprintf(w, "%s %d\n", level, s.counts[level]); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "TOTAL %d\nSKIPPED %d\n", s.total, s.skipped)
	return err
}
