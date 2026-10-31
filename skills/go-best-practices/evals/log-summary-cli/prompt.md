---
tags: [go-best-practices, cli]
max_turns: 30
timeout_seconds: 900
---

Write a practical, dependency-free Go CLI named logsum. It takes one directory argument
and reads the non-directory entries ending in .jsonl. Each nonblank line must be a JSON
object with a nonempty string level and a string ts. Other fields are fine. Invalid
records and blank lines are counted as skipped. Records up to 1 MiB must work.

Print lines as "LEVEL COUNT", ordered by count descending and then level ascending,
followed by "TOTAL N" and "SKIPPED N". An empty directory prints just the last two lines.
An unreadable entry should name its path on stderr, continue processing readable entries,
and produce a nonzero exit code. A missing directory or wrong argument count also returns
a nonzero exit code with a useful diagnostic. Include tests. A go.mod is already present.
