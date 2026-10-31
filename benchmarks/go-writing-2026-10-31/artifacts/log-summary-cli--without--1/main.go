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
	counts := make(map[string]int)
	total, skipped, status := 0, 0, 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		path := filepath.Join(args[0], entry.Name())
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(errOut, "logsum: %s: %v\n", path, err)
			status = 1
			continue
		}
		n, s, readErr := summarize(file, counts)
		total += n
		skipped += s
		closeErr := file.Close()
		for _, err := range []error{readErr, closeErr} {
			if err != nil {
				fmt.Fprintf(errOut, "logsum: %s: %v\n", path, err)
				status = 1
			}
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
		fmt.Fprintf(out, "%s %d\n", level, counts[level])
	}
	fmt.Fprintf(out, "TOTAL %d\nSKIPPED %d\n", total, skipped)
	return status
}

func summarize(input io.Reader, counts map[string]int) (total, skipped int, err error) {
	reader := bufio.NewReader(input)
	for {
		// ReadBytes avoids Scanner's default 64 KiB record limit, including
		// for a final record without a newline.
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			var record map[string]json.RawMessage
			var level, ts *string
			if len(bytes.TrimSpace(line)) == 0 || json.Unmarshal(line, &record) != nil ||
				json.Unmarshal(record["level"], &level) != nil || level == nil || *level == "" ||
				json.Unmarshal(record["ts"], &ts) != nil || ts == nil {
				skipped++
			} else {
				counts[*level]++
				total++
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return total, skipped, nil
			}
			return total, skipped, readErr
		}
	}
}
