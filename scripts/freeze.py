#!/usr/bin/env python3
"""Deliberately version evaluation inputs. Never called automatically by checks."""
import json
from common import ROOT, cases, digest

files = {str(p.relative_to(ROOT)): digest(p) for case in cases()
         for p in sorted(case.parent.rglob('*'))
         if p.is_file() and '__pycache__' not in p.parts}
(ROOT / 'evals.lock.json').write_text(json.dumps({'schema_version': 1, 'files': files}, indent=2) + '\n')
print(f'Frozen {len(files)} files. Treat a subsequent change as a new evaluation revision.')
