from __future__ import annotations

import subprocess
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
ACTION_SRC = REPO_ROOT / "actions" / "open-gitops-pr" / "src"
import sys

sys.path.insert(0, str(ACTION_SRC))

from gitops_pr_action.open_gitops_pr import (  # noqa: E402
    append_github_outputs,
    build_branch_name,
    load_targets,
    patch_manifest_image,
    process_target,
    select_targets,
)


class GitopsPrActionTests(unittest.TestCase):
    def test_load_targets_requires_unique_names(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir:
            config_path = Path(tmp_dir) / "gitops-targets.json"
            config_path.write_text(
                """
{
  "targets": [
    {
      "name": "demo",
      "gitops_repo": "wesen/repo",
      "gitops_branch": "main",
      "manifest_path": "gitops/app.yaml",
      "container_name": "demo"
    },
    {
      "name": "demo",
      "gitops_repo": "wesen/repo",
      "gitops_branch": "main",
      "manifest_path": "gitops/app2.yaml",
      "container_name": "demo"
    }
  ]
}
""".strip(),
                encoding="utf-8",
            )
            with self.assertRaisesRegex(ValueError, "duplicate target name"):
                load_targets(config_path)

    def test_select_targets_returns_named_target(self) -> None:
        targets = [
            {
                "name": "a",
                "gitops_repo": "wesen/repo",
                "gitops_branch": "main",
                "manifest_path": "gitops/app.yaml",
                "container_name": "a",
            },
            {
                "name": "b",
                "gitops_repo": "wesen/repo",
                "gitops_branch": "main",
                "manifest_path": "gitops/b.yaml",
                "container_name": "b",
            },
        ]
        self.assertEqual(select_targets(targets, "b", False), [targets[1]])

    def test_patch_manifest_image_updates_matching_container(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir:
            manifest_path = Path(tmp_dir) / "deployment.yaml"
            manifest_path.write_text(
                """
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: demo
          image: ghcr.io/wesen/demo:old
""".lstrip(),
                encoding="utf-8",
            )
            changed, original, updated = patch_manifest_image(
                manifest_path,
                container_name="demo",
                image="ghcr.io/wesen/demo:sha-1234567",
            )
            self.assertTrue(changed)
            self.assertIn("ghcr.io/wesen/demo:old", original)
            self.assertIn("ghcr.io/wesen/demo:sha-1234567", updated)

    def test_patch_manifest_image_noop_when_image_matches(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir:
            manifest_path = Path(tmp_dir) / "deployment.yaml"
            manifest_path.write_text(
                """
spec:
  template:
    spec:
      containers:
        - name: demo
          image: ghcr.io/wesen/demo:sha-1234567
""".lstrip(),
                encoding="utf-8",
            )
            changed, _, _ = patch_manifest_image(
                manifest_path,
                container_name="demo",
                image="ghcr.io/wesen/demo:sha-1234567",
            )
            self.assertFalse(changed)

    def test_patch_manifest_image_updates_only_named_container(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir:
            manifest_path = Path(tmp_dir) / "deployment.yaml"
            manifest_path.write_text(
                """
spec:
  template:
    spec:
      containers:
        - name: sidecar
          image: ghcr.io/wesen/sidecar:old
        - name: demo
          image: ghcr.io/wesen/demo:old
""".lstrip(),
                encoding="utf-8",
            )
            _, _, updated = patch_manifest_image(
                manifest_path,
                container_name="demo",
                image="ghcr.io/wesen/demo:sha-1234567",
            )
            self.assertIn("ghcr.io/wesen/sidecar:old", updated)
            self.assertIn("ghcr.io/wesen/demo:sha-1234567", updated)

    def test_build_branch_name_uses_image_and_target(self) -> None:
        branch = build_branch_name(
            image="ghcr.io/wesen/demo-app:sha-1234567",
            target_name="prod",
        )
        self.assertEqual(branch, "automation/demo-app-prod-sha-1234567")

    def test_process_target_dry_run_restores_manifest(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir:
            repo_dir = Path(tmp_dir) / "gitops"
            repo_dir.mkdir()
            subprocess.run(["git", "init"], cwd=repo_dir, check=True, capture_output=True, text=True)

            manifest_rel = Path("gitops/kustomize/demo/deployment.yaml")
            manifest_path = repo_dir / manifest_rel
            manifest_path.parent.mkdir(parents=True)
            original = """
spec:
  template:
    spec:
      containers:
        - name: demo
          image: ghcr.io/wesen/demo:old
""".lstrip()
            manifest_path.write_text(original, encoding="utf-8")

            result = process_target(
                target={
                    "name": "demo-prod",
                    "gitops_repo": "wesen/2026-03-27--hetzner-k3s",
                    "gitops_branch": "main",
                    "manifest_path": str(manifest_rel),
                    "container_name": "demo",
                },
                image="ghcr.io/wesen/demo:sha-1234567",
                gitops_repo_dir=repo_dir,
                dry_run=True,
                push=False,
                open_pr=False,
                token="",
            )

            self.assertTrue(result.changed)
            self.assertEqual(result.branch_name, "automation/demo-demo-prod-sha-1234567")
            self.assertEqual(manifest_path.read_text(encoding="utf-8"), original)

    def test_append_github_outputs_writes_machine_readable_fields(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir:
            output_path = Path(tmp_dir) / "github-output.txt"
            append_github_outputs(
                str(output_path),
                results=[],
            )
            self.assertEqual(
                output_path.read_text(encoding="utf-8"),
                "changed=false\nchanged_targets=\nbranch_names=\npr_numbers=\n",
            )


if __name__ == "__main__":
    unittest.main()
