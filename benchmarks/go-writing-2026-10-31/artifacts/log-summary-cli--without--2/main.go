// Command logsum summarizes the levels in a directory of JSONL logs.
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

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, errOut io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(errOut, "usage: logsum DIRECTORY")
		return 1
	}
	entries, err := os.ReadDir(args[0])
	if err != nil {
		fmt.Fprintf(errOut, "logsum: %s: %v\n", args[0], err)
		return 1
	}
	counts := make(map[string]uint64)
	var total, skipped uint64
	status := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		path := filepath.Join(args[0], entry.Name())
		if err := readFile(path, counts, &total, &skipped); err != nil {
			fmt.Fprintf(errOut, "logsum: %s: %v\n", path, err)
			status = 1
		}
	}
	levels := make([]string, 0, len(counts))
	for level := range counts {
		levels = append(levels, level)
	}
	sort.Slice(levels, func(i, j int) bool {
		if counts[levels[i]] != counts[levels[j]] {
			return counts[levels[i]] > counts[levels[j]]
		}
		return levels[i] < levels[j]
	})
	for _, level := range levels {
		if _, err := fmt.Fprintf(out, "%s %d\n", level, counts[level]); err != nil {
			fmt.Fprintf(errOut, "logsum: writing output: %v\n", err)
			return 1
		}
	}
	if _, err := fmt.Fprintf(out, "TOTAL %d\nSKIPPED %d\n", total, skipped); err != nil {
		fmt.Fprintf(errOut, "logsum: writing output: %v\n", err)
		return 1
	}
	return status
}

func readFile(path string, counts map[string]uint64, total, skipped *uint64) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	// ReadBytes avoids Scanner's default 64 KiB token limit, including for
	// records of exactly 1 MiB followed by a newline.
	r := bufio.NewReader(f)
	for {
		line, readErr := r.ReadBytes('\n')
		if len(line) > 0 && (readErr == nil || readErr == io.EOF) {
			level, valid := recordLevel(line)
			if valid {
				counts[level]++
				*total++
			} else {
				*skipped++
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

func recordLevel(line []byte) (string, bool) {
	if len(bytes.TrimSpace(line)) == 0 {
		return "", false
	}
	var record map[string]json.RawMessage
	if json.Unmarshal(line, &record) != nil || record == nil {
		return "", false
	}
	var level, ts *string
	if json.Unmarshal(record["level"], &level) != nil || level == nil || *level == "" {
		return "", false
	}
	if json.Unmarshal(record["ts"], &ts) != nil || ts == nil {
		return "", false
	}
	return *level, true
}
