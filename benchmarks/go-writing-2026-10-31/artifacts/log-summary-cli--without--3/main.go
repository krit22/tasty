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

type summary struct {
	counts  map[string]int
	total   int
	skipped int
}

func (s *summary) read(r io.Reader) error {
	reader := bufio.NewReader(r)
	for {
		// ReadBytes avoids Scanner's token limit, including for 1 MiB records.
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if len(line) > 0 {
			var record map[string]json.RawMessage
			var level, ts *string
			if len(bytes.TrimSpace(line)) == 0 || json.Unmarshal(line, &record) != nil || json.Unmarshal(record["level"], &level) != nil || level == nil || *level == "" || json.Unmarshal(record["ts"], &ts) != nil || ts == nil {
				s.skipped++
			} else {
				s.counts[*level]++
				s.total++
			}
		}
		if err == io.EOF {
			return nil
		}
	}
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: logsum DIRECTORY")
		return 1
	}
	entries, err := os.ReadDir(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "logsum: %s: %v\n", args[0], err)
		return 1
	}
	s := summary{counts: make(map[string]int)}
	status := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		path := filepath.Join(args[0], entry.Name())
		file, err := os.Open(path)
		if err == nil {
			err = s.read(file)
			if closeErr := file.Close(); err == nil {
				err = closeErr
			}
		}
		if err != nil {
			fmt.Fprintf(stderr, "logsum: %s: %v\n", path, err)
			status = 1
		}
	}
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
	var output bytes.Buffer
	for _, level := range levels {
		fmt.Fprintf(&output, "%s %d\n", level, s.counts[level])
	}
	fmt.Fprintf(&output, "TOTAL %d\nSKIPPED %d\n", s.total, s.skipped)
	if _, err := io.Copy(stdout, &output); err != nil {
		fmt.Fprintf(stderr, "logsum: writing output: %v\n", err)
		return 1
	}
	return status
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
