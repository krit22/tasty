package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func countFile(path string, counts map[string]int) (skipped int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 2*1024*1024)
	for scanner.Scan() {
		var record struct {
			Level string  `json:"level"`
			TS    *string `json:"ts"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil || record.Level == "" || record.TS == nil {
			skipped++
			continue
		}
		counts[record.Level]++
	}
	return skipped, scanner.Err()
}

func run(args []string, out, errOut io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(errOut, "usage: logsum DIRECTORY")
		return 1
	}
	entries, err := os.ReadDir(args[0])
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	counts := make(map[string]int)
	totalSkipped, status := 0, 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		path := filepath.Join(args[0], entry.Name())
		skipped, err := countFile(path, counts)
		totalSkipped += skipped
		if err != nil {
			fmt.Fprintf(errOut, "%s: %v\n", path, err)
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
	total := 0
	for _, level := range levels {
		total += counts[level]
		if _, err := fmt.Fprintf(out, "%s %d\n", level, counts[level]); err != nil {
			fmt.Fprintln(errOut, err)
			return 1
		}
	}
	if _, err := fmt.Fprintf(out, "TOTAL %d\nSKIPPED %d\n", total, totalSkipped); err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	return status
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
