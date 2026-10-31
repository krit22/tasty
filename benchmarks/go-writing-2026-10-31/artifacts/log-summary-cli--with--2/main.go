// Command logsum summarizes levels in JSON Lines files in a directory.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: logsum DIRECTORY")
		return 1
	}
	entries, err := os.ReadDir(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "logsum: read directory %s: %v\n", args[0], err)
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
			fmt.Fprintf(stderr, "logsum: %s: %v\n", path, err)
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
	var output strings.Builder
	for _, level := range levels {
		fmt.Fprintf(&output, "%s %d\n", level, counts[level])
	}
	fmt.Fprintf(&output, "TOTAL %d\nSKIPPED %d\n", total, skipped)
	if _, err := io.WriteString(stdout, output.String()); err != nil {
		fmt.Fprintf(stderr, "logsum: write output: %v\n", err)
		return 1
	}
	return status
}

func readFile(path string, counts map[string]uint64, total, skipped *uint64) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	return readRecords(file, counts, total, skipped)
}

func readRecords(input io.Reader, counts map[string]uint64, total, skipped *uint64) error {
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			level, valid := recordLevel(line)
			if valid {
				counts[level]++
				(*total)++
			} else {
				(*skipped)++
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

func recordLevel(line []byte) (string, bool) {
	if len(bytes.TrimSpace(line)) == 0 {
		return "", false
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(line, &fields); err != nil {
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
