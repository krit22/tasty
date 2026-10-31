"""Shared repository paths and reproducible local Go execution."""
from __future__ import annotations

import hashlib
import json
import os
from pathlib import Path
import signal
import shutil
import subprocess

ROOT = Path(__file__).resolve().parents[1]


def skill_dirs(root=ROOT):
    return sorted(p.parent for p in (root / 'skills').glob('*/SKILL.md'))


def cases(root=ROOT):
    return sorted((root / 'skills').glob('*/evals/*/case.yaml'))


def load_case(path):
    # JSON is a YAML subset: keep the case files readable without a YAML dependency.
    return json.loads(Path(path).read_text())


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def go_binary():
    value = os.environ.get('TASTY_GO', 'go')
    binary = shutil.which(value)
    if not binary:
        raise RuntimeError('Go not found; install Go or set TASTY_GO to its executable')
    return str(Path(binary).resolve())


def go_env():
    env = dict(os.environ)
    cache = Path(env.get('TASTY_CACHE', '/tmp/tasty-go-cache'))
    env.update(GOCACHE=str(cache / 'build'), GOMODCACHE=str(cache / 'modules'),
               GOTOOLCHAIN='local', GOWORK='off', GOFLAGS='-buildvcs=false')
    env['PATH'] = str(Path(go_binary()).parent) + os.pathsep + env['PATH']
    return env


def run(command, cwd, timeout=180, env=None):
    process = subprocess.Popen(command, cwd=cwd, env=env, stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE, text=True, start_new_session=True)
    try:
        out, err = process.communicate(timeout=timeout)
        return process.returncode, out, err
    except subprocess.TimeoutExpired:
        # Agent launchers and builds can have children; end the whole owned group.
        os.killpg(process.pid, signal.SIGTERM)
        try:
            out, err = process.communicate(timeout=5)
        except subprocess.TimeoutExpired:
            os.killpg(process.pid, signal.SIGKILL)
            out, err = process.communicate()
        return 124, out, err + '\nTimed out'


def payload_files(root=ROOT):
    files = [root / name for name in ('VERSION', 'LICENSE', 'PRIVACY.md')]
    for folder in skill_dirs(root):
        files.extend(p for p in folder.rglob('*') if p.is_file()
                     and not p.is_symlink()
                     and not {'evals', '__pycache__'}.intersection(p.relative_to(folder).parts))
    return sorted(files)
