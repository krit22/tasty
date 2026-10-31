#!/usr/bin/env python3
"""Run isolated task comparisons and grade behavior with withheld contract tests.

The checked-in fixtures and rubric are frozen by evals.lock.json. `prove` checks that
known-good implementations pass and specific deliberate regressions are detected.
`run` makes real model calls; `grade` only executes local candidate code.
"""
from __future__ import annotations

import argparse
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import random
import re
import shutil
import sys
import tempfile
import time

from common import ROOT, cases, digest, go_binary, go_env, load_case, run


def case_paths(root=ROOT):
    return {load_case(p)['name']: p.parent for p in cases(root)}


def copy_sources(source, target):
    """Copy ordinary Go inputs, not skills, grader files, logs, or symlink targets."""
    for path in sorted(Path(source).rglob('*')):
        relative = path.relative_to(source)
        if any(part.startswith('.') or part in ('vendor', 'node_modules') for part in relative.parts):
            continue
        if path.is_symlink() or not path.is_file():
            continue
        if path.suffix != '.go' and path.name not in ('go.mod', 'go.sum'):
            continue
        destination = Path(target) / relative
        destination.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(path, destination)


def grade(case_dir, workspace, race=False):
    spec = load_case(case_dir / 'case.yaml')
    with tempfile.TemporaryDirectory(prefix='tasty-grade-') as tmp:
        target = Path(tmp)
        copy_sources(Path(workspace), target)
        env = go_env()
        binary = target / 'tasty-app'
        command = [go_binary(), 'build']
        command += ['-o', str(binary), '.'] if spec['kind'] == 'cli' else ['./...']
        status, out, err = run(command, target, env=env)
        build = {'passed': status == 0, 'exit_code': status, 'evidence': (out+err)[-4000:]}
        if status:
            return {'case': spec['name'], 'build': build, 'checks': [
                dict(c, passed=False, evidence='candidate did not build') for c in spec['checks']],
                'manual': 'ungraded'}
        for source in (case_dir / 'validation').glob('*_test.go'):
            shutil.copyfile(source, target / ('zz_tasty_' + source.name))
        env['TASTY_EVAL_BINARY'] = str(binary)
        command = [go_binary(), 'test', '-json', '-count=1', '-timeout=30s']
        if race:
            command.append('-race')
        command.append('./...')
        status, out, err = run(command, target, env=env)
        test_events = {}
        for line in out.splitlines():
            try:
                event = json.loads(line)
            except ValueError:
                continue
            test = event.get('Test', '')
            if test:
                test_events.setdefault(test, []).append(event)
        results = []
        for check in spec['checks']:
            events = test_events.get(check['test'], [])
            passed = any(e.get('Action') == 'pass' for e in events)
            evidence = ''.join(e.get('Output', '') for e in events)[-2500:]
            if not events:
                evidence = f'No completed result for test; go test exited {status}.\n{out[-1500:]}{err[-500:]}'
            results.append(dict(check, passed=passed, evidence=evidence))
        return {'case': spec['name'], 'build': build, 'tests_exit_code': status,
                'checks': results, 'manual': 'ungraded', 'race_enabled': race}


def reference_workspace(case_dir, target):
    shutil.copytree(case_dir / 'files', target, dirs_exist_ok=True)
    shutil.copytree(case_dir / 'reference', target, dirs_exist_ok=True)


def prove(selected):
    errors = []
    for name, folder in selected.items():
        with tempfile.TemporaryDirectory(prefix='tasty-reference-') as tmp:
            workspace = Path(tmp)
            reference_workspace(folder, workspace)
            result = grade(folder, workspace)
            failures = [c['id'] for c in result['checks'] if not c['passed']]
            if failures:
                errors.append(f'{name}: known-good implementation failed {failures}: {result}')
            print(f'{name}: reference {len(result["checks"])-len(failures)}/{len(result["checks"])}')
        for mutation in json.loads((folder / 'mutations.json').read_text()):
            with tempfile.TemporaryDirectory(prefix='tasty-mutant-') as tmp:
                workspace = Path(tmp)
                reference_workspace(folder, workspace)
                file = workspace / mutation['file']
                text = file.read_text()
                if text.count(mutation['before']) != 1:
                    errors.append(f'{name}/{mutation["name"]}: mutation does not match exactly once')
                    continue
                file.write_text(text.replace(mutation['before'], mutation['after'], 1))
                result = grade(folder, workspace)
                detected = result['build']['passed'] and any(
                    c['id'] == mutation['check'] and not c['passed'] for c in result['checks'])
                print(f'{name}: mutation {mutation["name"]}: {"detected" if detected else "MISSED"}')
                if not detected:
                    errors.append(f'{name}: regression {mutation["name"]} not detected by {mutation["check"]}')
    return errors


def prompt_body(path):
    return path.read_text().split('---', 2)[2].strip()


def execute_run(job):
    folder, condition, repetition, destination, timeout = job
    name = folder.name
    run_dir = destination / f'{name}--{condition}--{repetition}'
    if run_dir.exists():
        raise RuntimeError(f'refusing to overwrite existing run: {run_dir}')
    workspace = run_dir / 'workspace'
    workspace.mkdir(parents=True)
    shutil.copytree(folder / 'files', workspace, dirs_exist_ok=True)
    skill_hashes = {}
    if condition == 'with':
        skill = ROOT / 'skills/go-best-practices'
        installed = workspace / '.agents/skills/go-best-practices'
        for path in sorted(skill.rglob('*')):
            if not path.is_file() or 'evals' in path.relative_to(skill).parts:
                continue
            target = installed / path.relative_to(skill)
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, target)
            skill_hashes[str(path.relative_to(ROOT))] = digest(path)
    preamble = (
        'Work only in the current workspace. Go is at ' + go_binary() + '. '
        'Set PATH to include its bin directory if needed. Use the provided GOCACHE and '
        'GOMODCACHE. All inputs are local; no external dependencies are needed. '
        'Do not inspect parent or neighboring directories.\n\n'
    )
    if condition == 'with':
        preamble += ('Read .agents/skills/go-best-practices/SKILL.md before starting and follow '
                     'it; read its references when relevant.\n\n')
    else:
        preamble += 'Tasty is not installed in this workspace. Complete the task normally.\n\n'
    prompt = preamble + prompt_body(folder / 'prompt.md')
    (run_dir / 'prompt.txt').write_text(prompt)
    command = ['codex', 'exec', '--ignore-user-config', '--ephemeral', '--skip-git-repo-check',
               '--sandbox', 'workspace-write', '--json', '-C', str(workspace),
               '-o', str(run_dir / 'last-message.txt'), prompt]
    started = datetime.now(timezone.utc).isoformat()
    begin = time.monotonic()
    status, out, err = run(command, workspace, timeout=timeout, env=go_env())
    (run_dir / 'events.jsonl').write_text(out)
    (run_dir / 'stderr.txt').write_text(err)
    completed = any('"type":"turn.completed"' in line.replace(' ', '') for line in out.splitlines())
    usage = None
    for line in out.splitlines():
        try:
            event = json.loads(line)
            if event.get('type') == 'turn.completed':
                usage = event.get('usage')
        except ValueError:
            pass
    metadata = {'case': name, 'condition': condition, 'repetition': repetition,
                'started': started, 'seconds': round(time.monotonic()-begin, 2),
                'agent_exit_code': status, 'status': 'completed' if completed and status == 0 else 'runner_error',
                'usage': usage, 'skill_hashes': skill_hashes,
                'evaluation_lock_sha256': digest(ROOT / 'evals.lock.json')}
    if metadata['status'] == 'completed':
        metadata['grade'] = grade(folder, workspace)
    (run_dir / 'result.json').write_text(json.dumps(metadata, indent=2) + '\n')
    print(f'{run_dir.name}: {metadata["status"]}', flush=True)
    return metadata


def summarize(directory):
    result = {'runs': [], 'totals': {}, 'scope': 'automated behavioral checks only; review rubrics ungraded'}
    locks = set()
    for path in sorted(Path(directory).glob('*/result.json')):
        data = json.loads(path.read_text())
        locks.add(data['evaluation_lock_sha256'])
        result['runs'].append(data)
        row = result['totals'].setdefault(data['condition'], {'completed': 0, 'runner_errors': 0, 'passed': 0, 'checks': 0, 'builds_passed': 0, 'test_suites_passed': 0})
        if data['status'] != 'completed':
            row['runner_errors'] += 1
            continue
        row['completed'] += 1
        row['builds_passed'] += int(data['grade']['build']['passed'])
        row['test_suites_passed'] += int(data['grade'].get('tests_exit_code', 1) == 0)
        row['passed'] += sum(c['passed'] for c in data['grade']['checks'])
        row['checks'] += len(data['grade']['checks'])
    if len(locks) > 1:
        raise RuntimeError('cannot aggregate runs from different frozen evaluation revisions')
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('action', choices=['prove', 'grade', 'run', 'summarize'])
    parser.add_argument('--case', choices=sorted(case_paths()))
    parser.add_argument('--workspace', type=Path)
    parser.add_argument('--out', type=Path)
    parser.add_argument('--repetitions', type=int, default=3)
    parser.add_argument('--parallel', type=int, default=2)
    parser.add_argument('--timeout', type=int, default=900)
    parser.add_argument('--race', action='store_true')
    args = parser.parse_args()
    selected = case_paths()
    if args.case:
        selected = {args.case: selected[args.case]}
    if args.action == 'prove':
        errors = prove(selected)
        for error in errors:
            print(error, file=sys.stderr)
        return bool(errors)
    if args.action == 'grade':
        if not args.case or not args.workspace:
            parser.error('grade requires --case and --workspace')
        data = grade(selected[args.case], args.workspace, args.race)
    elif args.action == 'run':
        from check import check
        problems = check()
        if problems:
            raise RuntimeError('validation failed before model calls: ' + '; '.join(problems))
        if not args.out:
            parser.error('run requires --out for immutable run artifacts')
        if args.repetitions < 1 or args.parallel < 1:
            parser.error('repetitions and parallel must be positive')
        destination = args.out.resolve()
        destination.mkdir(parents=True, exist_ok=True)
        if any(destination.iterdir()):
            raise RuntimeError('run destination must be empty; never overwrite prior evidence')
        env = go_env()
        _, go_version, _ = run([go_binary(), 'version'], ROOT, env=env)
        status, agent_version, err = run(['codex', '--version'], ROOT, env=env)
        if status:
            raise RuntimeError('codex runner unavailable: ' + err)
        (destination / 'environment.json').write_text(json.dumps({
            'go': go_version.strip(), 'agent': agent_version.strip(),
            'model': 'runner default; inspect events for any reported model identity',
            'settings': 'ignore-user-config; workspace-write; ephemeral; explicit skill loading',
            'evaluates_discovery': False,
            'evaluation_lock_sha256': digest(ROOT / 'evals.lock.json'),
        }, indent=2) + '\n')
        jobs = [(folder, condition, n, destination, args.timeout) for folder in selected.values()
                for condition in ('without', 'with') for n in range(1, args.repetitions+1)]
        random.Random(20261031).shuffle(jobs)
        with ThreadPoolExecutor(max_workers=args.parallel) as executor:
            list(executor.map(execute_run, jobs))
        data = summarize(destination)
        (destination / 'summary.json').write_text(json.dumps(data, indent=2) + '\n')
    else:
        if not args.out:
            parser.error('summarize requires --out pointing at a run directory')
        data = summarize(args.out)
    print(json.dumps(data, indent=2))
    return 0


if __name__ == '__main__':
    try:
        sys.exit(main())
    except (RuntimeError, FileNotFoundError) as exc:
        sys.exit(str(exc))
