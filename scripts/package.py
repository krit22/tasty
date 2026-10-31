#!/usr/bin/env python3
"""Build reproducible per-skill ZIPs, excluding evaluations and machine output."""
import argparse
from pathlib import Path
import zipfile

from common import ROOT, payload_files, skill_dirs


def build(out, root=ROOT):
    out = Path(out)
    out.mkdir(parents=True, exist_ok=True)
    archives = []
    version = (root / 'VERSION').read_text().strip()
    for skill in skill_dirs(root):
        archive = out / f'{skill.name}-{version}.zip'
        with zipfile.ZipFile(archive, 'w', zipfile.ZIP_DEFLATED) as z:
            for file in payload_files(root):
                if file.is_relative_to(root / 'skills'):
                    if not file.is_relative_to(skill):
                        continue
                    target = Path(skill.name) / file.relative_to(skill)
                else:
                    target = Path(skill.name) / file.name
                info = zipfile.ZipInfo(target.as_posix(), (1980, 1, 1, 0, 0, 0))
                info.compress_type = zipfile.ZIP_DEFLATED
                info.external_attr = 0o644 << 16
                z.writestr(info, file.read_bytes())
        archives.append(archive)
    return archives


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--out', type=Path, default=ROOT / 'dist')
    args = parser.parse_args()
    for path in build(args.out):
        print(path)
