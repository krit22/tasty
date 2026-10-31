#!/usr/bin/env python3
"""Compile and execute every complete Go documentation example."""
from __future__ import annotations

import re
import sys
import tempfile
from pathlib import Path

from common import ROOT, go_binary, go_env, run


def examples(root=ROOT):
    docs = sorted((root / 'skills').glob('*/SKILL.md'))
    docs += sorted((root / 'skills').glob('*/references/*.md'))
    found = []
    for path in docs:
        for number, match in enumerate(re.finditer(r'^```go\n(.*?)^```\s*$', path.read_text(), re.M | re.S), 1):
            found.append((path, number, match[1]))
    return found


def check(root=ROOT):
    failures = []
    for doc, number, source in examples(root):
        with tempfile.TemporaryDirectory(prefix='tasty-example-') as tmp:
            path = Path(tmp)
            (path / 'go.mod').write_text('module example.com/tasty-example\n\ngo 1.26\n')
            (path / 'example_test.go').write_text(source)
            status, out, err = run([go_binary(), 'test', '-count=1', './...'], path, env=go_env())
            if status:
                failures.append(f'{doc}:{number}: {out}{err}')
    return failures


if __name__ == '__main__':
    try:
        failures = check()
        for failure in failures:
            print(failure, file=sys.stderr)
        print(f'{len(examples())} examples executed; {len(failures)} failed; 0 skipped')
        sys.exit(bool(failures))
    except RuntimeError as exc:
        sys.exit(str(exc))
