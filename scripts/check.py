#!/usr/bin/env python3
"""Validate skill structure, source traces, and frozen evaluation inputs."""
from __future__ import annotations

import json
from pathlib import Path
import re
import sys

from common import ROOT, cases, digest, load_case, skill_dirs


def check(root=ROOT):
    problems = []
    sources = json.loads((root / 'sources/sources.json').read_text())
    source_ids = {s['id'] for s in sources}
    for source in sources:
        if not source['url'].startswith('https://') or not source.get('reviewed'):
            problems.append(f"source {source['id']}: missing URL or review date")
    rules = json.loads((root / 'sources/rules.json').read_text())
    rule_ids = [r['id'] for r in rules]
    if len(rule_ids) != len(set(rule_ids)):
        problems.append('duplicate rule identifiers')
    for rule in rules:
        if not rule.get('reason') or not rule.get('exceptions') or not rule.get('verification'):
            problems.append(f"rule {rule['id']}: incomplete rationale, exceptions, or verification")
        if not rule.get('sources') or set(rule['sources']) - source_ids:
            problems.append(f"rule {rule['id']}: unknown or missing source")
    folders = skill_dirs(root)
    if not folders:
        problems.append('no skills found')
    for folder in folders:
        path = folder / 'SKILL.md'
        content = path.read_text()
        match = re.match(r'\A---\n(.*?)\n---\n(.*)', content, re.S)
        if not match:
            problems.append(f'{path}: missing frontmatter')
            continue
        fields = dict(re.findall(r'^([a-z-]+): (.+)$', match[1], re.M))
        if fields.get('name') != folder.name or len(folder.name) > 64 or not re.fullmatch(r'[a-z0-9]+(?:-[a-z0-9]+)*', folder.name):
            problems.append(f'{path}: invalid name')
        if not 1 <= len(fields.get('description', '')) <= 1024:
            problems.append(f'{path}: invalid description')
        if len(match[2].splitlines()) > 500:
            problems.append(f'{path}: move detail out of the entrypoint (over 500 lines)')
        for doc in [path, *folder.glob('references/*.md')]:
            text = doc.read_text()
            for url in re.findall(r'\[[^\]]*\]\(([^)]+)\)', text):
                if url.startswith(('https://', 'http://', '#')):
                    continue
                target = (doc.parent / url.split('#')[0]).resolve()
                if not target.is_relative_to(root.resolve()) or not target.is_file():
                    problems.append(f'{doc}: broken local link {url}')
            unknown = set(re.findall(r'\bGO-[A-Z]+(?:-[A-Z]+)*\b', text)) - set(rule_ids)
            if unknown:
                problems.append(f'{doc}: unknown rules {sorted(unknown)}')
        for ref in folder.glob('references/*.md'):
            if f'references/{ref.name}' not in content:
                problems.append(f'{ref}: not linked from SKILL.md')
    for path in cases(root):
        data = load_case(path)
        folder = path.parent
        if data['name'] != folder.name or data['schema_version'] != '1.1':
            problems.append(f'{path}: invalid name or schema')
        for name in ('prompt.md', 'place-files.sh'):
            if not (folder / name).is_file():
                problems.append(f'{folder}: missing {name}')
        checks = data.get('checks', [])
        if not checks or len({c['test'] for c in checks}) != len(checks):
            problems.append(f'{path}: missing or duplicate automated checks')
        for item in checks:
            if not re.fullmatch(r'TestTasty[A-Za-z0-9]+', item['test']):
                problems.append(f'{path}: invalid test {item["test"]}')
            if item['rule'] not in rule_ids:
                problems.append(f'{path}: unknown rule {item["rule"]}')
            tests = '\n'.join(p.read_text() for p in (folder / 'validation').glob('*_test.go'))
            if f'func {item["test"]}(' not in tests:
                problems.append(f'{path}: automated test {item["test"]} is absent')
        graders = sorted((folder / 'graders').glob('*.md'))
        if [int(p.name.split('-')[0]) for p in graders] != list(range(1, len(graders)+1)):
            problems.append(f'{folder}: grader numbering is not contiguous')
        if not graders:
            problems.append(f'{folder}: missing review rubrics')
    lock = json.loads((root / 'evals.lock.json').read_text())
    actual = {str(p.relative_to(root)): digest(p) for case in cases(root)
              for p in case.parent.rglob('*') if p.is_file() and '__pycache__' not in p.parts}
    if actual != lock['files']:
        changed = sorted(k for k in actual.keys() | lock['files'].keys()
                         if actual.get(k) != lock['files'].get(k))
        problems.append('frozen evaluations changed: ' + ', '.join(changed))
    return problems


if __name__ == '__main__':
    errors = check()
    for error in errors:
        print(error, file=sys.stderr)
    print(f'{len(errors)} errors; {len(skill_dirs())} skill(s); {len(cases())} frozen cases')
    sys.exit(bool(errors))
