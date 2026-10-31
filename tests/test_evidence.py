"""Publishing evidence must preserve generation records and grade published inputs."""
import json
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / 'scripts'))
from common import ROOT, digest
from export_results import export


class EvidenceTests(unittest.TestCase):
    def test_export_preserves_records_and_assesses_the_published_snapshot(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            runs = root / 'runs'; runs.mkdir()
            (runs / 'environment.json').write_text('{}')
            record_dir = runs / 'case--with--1'; workspace = record_dir / 'workspace'
            workspace.mkdir(parents=True)
            (workspace / 'go.mod').write_text('module example.com/semver\n\ngo 1.26\n')
            (workspace / 'semver.go').write_text('package semver\n')
            record = record_dir / 'result.json'
            record.write_text(json.dumps({'case':'semver-library-api','condition':'with',
                'status':'completed','seconds':1,'usage':None,
                'evaluation_lock_sha256':digest(ROOT / 'evals.lock.json')}))
            before = record.read_bytes()
            destination = root / 'published'

            def assess(case, candidate):
                self.assertTrue(candidate.is_relative_to(destination))
                self.assertEqual((candidate / 'semver.go').read_text(), 'package semver\n')
                return {'build':{'passed':True}, 'tests_exit_code':0,
                        'checks':[{'passed':True}], 'manual':'ungraded'}

            with patch('export_results.grade', side_effect=assess), patch('export_results.run', return_value=(0,'go version test','')):
                result = export(runs, destination)
            self.assertEqual(record.read_bytes(), before)
            self.assertEqual(result['totals']['with']['passed'], 1)
            self.assertEqual(result['runs'][0]['grade']['manual'], 'ungraded')
            self.assertEqual(result['runs'][0]['artifact_sha256']['semver.go'],
                             digest(destination / 'artifacts/case--with--1/semver.go'))

    def test_export_never_overwrites_existing_evidence(self):
        with tempfile.TemporaryDirectory() as tmp:
            with self.assertRaisesRegex(RuntimeError, 'overwrite'):
                export(Path(tmp), Path(tmp))


if __name__ == '__main__':
    unittest.main()
