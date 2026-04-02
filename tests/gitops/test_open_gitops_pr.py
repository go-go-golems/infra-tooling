from __future__ import annotations

import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
ACTION_SRC = REPO_ROOT / "actions" / "open-gitops-pr" / "src"
import sys

sys.path.insert(0, str(ACTION_SRC))

from gitops_pr_action.open_gitops_pr import (  # noqa: E402
    build_branch_name,
    load_targets,
    patch_manifest_image,
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

    def test_build_branch_name_uses_image_and_target(self) -> None:
        branch = build_branch_name(
            image="ghcr.io/wesen/demo-app:sha-1234567",
            target_name="prod",
        )
        self.assertEqual(branch, "automation/demo-app-prod-sha-1234567")


if __name__ == "__main__":
    unittest.main()
