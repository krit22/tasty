"""Meaningful checks for evaluation isolation, drift detection, and packaging."""
from pathlib import Path
import json
import os
import shutil
import sys
import tempfile
import unittest
from unittest.mock import patch
import zipfile

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / 'scripts'))
from common import ROOT, payload_files, run
import check
import evaluate
import package


class ToolingTests(unittest.TestCase):
    def test_timed_out_commands_are_reported_without_hanging(self):
        status, out, err = run([sys.executable, '-c', 'import time; time.sleep(20)'], ROOT, timeout=0.05)
        self.assertEqual(status, 124)
        self.assertIn('Timed out', err)

    def test_repository_is_valid(self):
        self.assertEqual(check.check(), [])

    def test_payload_excludes_planted_bugs_and_reference_solutions(self):
        files = payload_files()
        self.assertTrue(any(p.name == 'SKILL.md' for p in files))
        self.assertFalse(any('evals' in p.relative_to(ROOT).parts for p in files))

    def test_package_is_reproducible_and_installable(self):
        with tempfile.TemporaryDirectory() as tmp:
            first = package.build(Path(tmp) / 'a')[0]
            second = package.build(Path(tmp) / 'b')[0]
            self.assertEqual(first.read_bytes(), second.read_bytes())
            with zipfile.ZipFile(first) as z:
                names = z.namelist()
                self.assertIn('go-best-practices/SKILL.md', names)
                self.assertIn('go-best-practices/LICENSE', names)
                self.assertTrue(any('/references/' in name for name in names))
                self.assertFalse(any('/evals/' in name or name.startswith('/') or '..' in Path(name).parts for name in names))

    def test_grading_copy_cannot_import_skills_or_symlink_targets(self):
        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp)
            source, target = base / 'input', base / 'output'
            source.mkdir(); target.mkdir()
            (source / 'main.go').write_text('package main')
            (source / 'go.mod').write_text('module x')
            secret = base / 'private.go'; secret.write_text('private')
            (source / 'linked.go').symlink_to(secret)
            hidden = source / '.agents'; hidden.mkdir()
            (hidden / 'hidden.go').write_text('not candidate code')
            (source / 'notes.md').write_text('not candidate code')
            evaluate.copy_sources(source, target)
            self.assertEqual(sorted(p.name for p in target.iterdir()), ['go.mod', 'main.go'])

    def test_missing_test_events_do_not_count_as_passing(self):
        folder = evaluate.case_paths()['semver-library-api']
        with tempfile.TemporaryDirectory() as tmp:
            evaluate.reference_workspace(folder, Path(tmp))
            with patch.object(evaluate, 'run', side_effect=[(0, '', ''), (0, '', '')]):
                result = evaluate.grade(folder, Path(tmp))
            self.assertFalse(any(c['passed'] for c in result['checks']))

    def test_build_failure_is_evidence_of_candidate_failure(self):
        folder = evaluate.case_paths()['semver-library-api']
        with tempfile.TemporaryDirectory() as tmp:
            evaluate.reference_workspace(folder, Path(tmp))
            with patch.object(evaluate, 'run', return_value=(1, '', 'invalid Go')):
                result = evaluate.grade(folder, Path(tmp))
            self.assertFalse(result['build']['passed'])
            self.assertFalse(any(c['passed'] for c in result['checks']))

    def test_runner_failure_is_not_a_model_score(self):
        with tempfile.TemporaryDirectory() as tmp:
            run = Path(tmp) / 'failed'; run.mkdir()
            (run / 'result.json').write_text(json.dumps({'condition':'without', 'status':'runner_error', 'evaluation_lock_sha256':'a'}))
            totals = evaluate.summarize(tmp)['totals']['without']
            self.assertEqual(totals['runner_errors'], 1)
            self.assertEqual(totals['checks'], 0)

    def test_mixed_evaluation_revisions_cannot_be_aggregated(self):
        with tempfile.TemporaryDirectory() as tmp:
            for n, lock in enumerate(['a', 'b']):
                run = Path(tmp) / str(n); run.mkdir()
                (run / 'result.json').write_text(json.dumps({'condition':'with', 'status':'runner_error', 'evaluation_lock_sha256':lock}))
            with self.assertRaisesRegex(RuntimeError, 'different frozen'):
                evaluate.summarize(tmp)

    def test_frozen_input_change_is_detected(self):
        with tempfile.TemporaryDirectory() as tmp:
            target = Path(tmp)
            shutil.copytree(ROOT / 'sources', target / 'sources')
            shutil.copytree(ROOT / 'skills', target / 'skills')
            shutil.copyfile(ROOT / 'evals.lock.json', target / 'evals.lock.json')
            prompt = next((target / 'skills').glob('*/evals/*/prompt.md'))
            prompt.write_text(prompt.read_text() + '\nChanged task.\n')
            self.assertTrue(any('frozen evaluations changed' in e for e in check.check(target)))


if __name__ == '__main__':
    unittest.main()
