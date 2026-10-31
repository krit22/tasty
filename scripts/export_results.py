#!/usr/bin/env python3
"""Assess completed outputs with the current grader and publish reproducible evidence.

Generation records under --runs are never modified. Published artifacts include only
the generated Go inputs, change notes, and final response, not host configuration or
raw tool-event logs. Qualitative rubrics remain explicitly ungraded.
"""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import re
import shutil

from common import ROOT, digest, go_binary, go_env, run
from evaluate import case_paths, copy_sources, grade


def export(runs, destination):
    if destination.exists():
        raise RuntimeError('refusing to overwrite published evidence; use a new directory')
    records = sorted(runs.glob('*/result.json'))
    if not records:
        raise RuntimeError('no completed run records found')
    destination.mkdir(parents=True)
    environment = json.loads((runs / 'environment.json').read_text())
    _, version, _ = run([go_binary(), 'version'], ROOT, env=go_env())
    results = {'environment': environment, 'assessment': {
        'go': version.strip(), 'GOFLAGS': '-buildvcs=false',
        'grader_sha256': digest(ROOT / 'scripts/evaluate.py'),
        'common_sha256': digest(ROOT / 'scripts/common.py'),
        'scope': 'automated behavioral checks; qualitative rubrics ungraded',
    }, 'runs': [], 'totals': {}}
    locks = {json.loads(record.read_text())['evaluation_lock_sha256'] for record in records}
    if locks != {digest(ROOT / 'evals.lock.json')}:
        raise RuntimeError('run records do not match the current frozen evaluation revision')
    suspect = re.compile(r'-----BEGIN [A-Z ]*PRIVATE KEY-----|gh[pousr]_[A-Za-z0-9]{20,}|sk-[A-Za-z0-9]{24,}')
    for record in records:
        metadata = json.loads(record.read_text())
        entry = {key: value for key, value in metadata.items() if key != 'grade'}
        entry['generation_record_sha256'] = digest(record)
        entry['run_id'] = record.parent.name
        if entry['status'] == 'completed':
            snapshot = destination / 'artifacts' / entry['run_id']
            snapshot.mkdir(parents=True)
            copy_sources(record.parent / 'workspace', snapshot)
            for source, name in [(record.parent / 'workspace/CHANGES.md', 'CHANGES.md'),
                                 (record.parent / 'last-message.txt', 'RESPONSE.txt')]:
                if source.is_file() and not source.is_symlink():
                    shutil.copyfile(source, snapshot / name)
            for file in snapshot.rglob('*'):
                if file.is_file() and suspect.search(file.read_text()):
                    raise RuntimeError(f'inspect possible credential before publishing: {file}')
            # Grade the exact published snapshot, so anyone can reproduce this assessment.
            entry['grade'] = grade(case_paths()[entry['case']], snapshot)
            entry['artifacts'] = str(snapshot.relative_to(destination))
            entry['artifact_sha256'] = {str(p.relative_to(snapshot)): digest(p)
                                       for p in sorted(snapshot.rglob('*')) if p.is_file()}
        results['runs'].append(entry)
        row = results['totals'].setdefault(entry['condition'], {
            'completed': 0, 'runner_errors': 0, 'passed': 0, 'checks': 0,
            'builds_passed': 0, 'test_suites_passed': 0, 'input_tokens': 0,
            'cached_input_tokens': 0, 'output_tokens': 0, 'seconds': 0,
        })
        if entry['status'] != 'completed':
            row['runner_errors'] += 1
            continue
        row['completed'] += 1
        row['builds_passed'] += int(entry['grade']['build']['passed'])
        row['test_suites_passed'] += int(entry['grade'].get('tests_exit_code', 1) == 0)
        row['passed'] += sum(c['passed'] for c in entry['grade']['checks'])
        row['checks'] += len(entry['grade']['checks'])
        row['seconds'] += entry['seconds']
        for key in ('input_tokens', 'cached_input_tokens', 'output_tokens'):
            row[key] += (entry.get('usage') or {}).get(key, 0)
    (destination / 'results.json').write_text(json.dumps(results, indent=2) + '\n')
    return results


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--runs', type=Path, required=True)
    parser.add_argument('--out', type=Path, required=True)
    args = parser.parse_args()
    try:
        data = export(args.runs.resolve(), args.out.resolve())
        print(json.dumps(data['totals'], indent=2))
    except RuntimeError as exc:
        raise SystemExit(str(exc))
