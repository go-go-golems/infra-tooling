# Playbook: Migrating Git worktrees to a new canonical clone

Use this when a repository used to live as a submodule or nested clone and has many worktrees under `~/workspaces`, but the canonical clone has moved elsewhere.

This playbook was written during the `glazed` migration from:

- old common dir: `/home/manuel/code/wesen/corporate-headquarters/.git/modules/glazed`
- new canonical clone: `/home/manuel/code/wesen/go-go-golems/glazed`

The same procedure applies to the other go-go-golems repositories that were removed from `corporate-headquarters`.

## Goal

Recreate clean worktrees so their `.git` files point at the new canonical repository, while preserving:

- branch names,
- HEAD commits,
- remote-pushed commits,
- dirty worktrees for later manual handling.

Git does not provide a direct “reparent this worktree to another common dir” command. The safe practical procedure is:

1. classify old worktrees,
2. push or otherwise preserve unpublished commits,
3. skip dirty worktrees,
4. copy branch refs into the new canonical clone,
5. remove the old worktree,
6. recreate it from the new canonical clone at the same path.

## Safety rules

Never bulk-migrate a worktree unless all of these are true:

- `git status --porcelain --untracked-files=all` is empty,
- `git rev-list --count HEAD --not --remotes` is `0`,
- the branch name is known,
- the target canonical clone has the commit before the old worktree is removed.

Leave dirty worktrees alone until their uncommitted work has been committed, stashed, or exported as patches.

## Discover old worktrees

For a package named `glazed`:

```bash
oldmeta=/home/manuel/code/wesen/corporate-headquarters/.git/modules/glazed
newrepo=/home/manuel/code/wesen/go-go-golems/glazed

git -C "$oldmeta" worktree list --porcelain
```

To find worktrees under `~/workspaces` whose `.git` file still points to the old common dir:

```bash
find /home/manuel/workspaces -name .git -type f -print |
while read -r gitfile; do
  if grep -q "$oldmeta" "$gitfile"; then
    dirname "$gitfile"
  fi
done
```

## Fetch before classifying

Fetch all remotes in the old common dir so “unpushed” detection is accurate:

```bash
for remote in $(git -C "$oldmeta" remote); do
  git -C "$oldmeta" fetch "$remote" --prune
done

git -C "$newrepo" fetch origin --prune
```

## Classify a worktree

For each worktree path `$wt`:

```bash
branch=$(git -C "$wt" branch --show-current)
head=$(git -C "$wt" rev-parse HEAD)
dirty=$(git -C "$wt" status --porcelain=v1 --untracked-files=all)
unpushed=$(git -C "$wt" rev-list --count HEAD --not --remotes)

printf 'branch=%s\nhead=%s\ndirty_lines=%s\nunpushed=%s\n' \
  "$branch" "$head" "$(printf '%s\n' "$dirty" | sed '/^$/d' | wc -l)" "$unpushed"
```

Interpretation:

- dirty lines > 0: skip for now,
- unpushed > 0: push or otherwise preserve first,
- both clean and unpushed = 0: safe to migrate.

## Push unpublished commits

Preferred when branches should survive machine loss or be reviewable on GitHub:

```bash
git -C "$wt" push -u origin "$branch"
```

If a pre-push hook blocks an intentional archival push, make the bypass explicit and record it:

```bash
git -C "$wt" push --no-verify -u origin "$branch"
```

During the `glazed` migration, one pre-push hook ran the full validation suite and failed on pre-existing `gosec` findings, so the archival pushes were repeated with `--no-verify` after confirming the worktrees were clean.

## Migrate a single clean worktree

Inputs:

```bash
oldmeta=/home/manuel/code/wesen/corporate-headquarters/.git/modules/glazed
newrepo=/home/manuel/code/wesen/go-go-golems/glazed
wt=/home/manuel/workspaces/YYYY-MM-DD/some-workspace/glazed
branch=task/some-branch
expected_head=$(git -C "$wt" rev-parse HEAD)
```

1. Re-check safety immediately before making changes:

```bash
test -z "$(git -C "$wt" status --porcelain=v1 --untracked-files=all)"
test "$(git -C "$wt" rev-list --count HEAD --not --remotes)" = 0
```

2. Copy the old local branch ref into the new canonical repo:

```bash
git -C "$newrepo" fetch "$oldmeta" \
  "+refs/heads/$branch:refs/heads/$branch"

test "$(git -C "$newrepo" rev-parse "$branch")" = "$expected_head"
```

3. Remove the old worktree from the old common dir:

```bash
git -C "$oldmeta" worktree remove --force "$wt"
```

4. Recreate it from the new canonical clone at the same path:

```bash
git -C "$newrepo" worktree add "$wt" "$branch"
```

5. Restore an upstream if the branch exists on `origin`:

```bash
if git -C "$wt" rev-parse --verify --quiet "origin/$branch" >/dev/null; then
  git -C "$wt" branch --set-upstream-to="origin/$branch" "$branch"
fi
```

6. Verify:

```bash
test "$(git -C "$wt" rev-parse HEAD)" = "$expected_head"
grep -q '/go-go-golems/glazed/.git/worktrees/' "$wt/.git"
git -C "$wt" status --short
```

## Bulk migration script pattern

This is the pattern used for clean `glazed` worktrees. Adapt `oldmeta`, `newrepo`, and the path filter for another repository.

```python
from pathlib import Path
import subprocess

oldmeta = Path('/home/manuel/code/wesen/corporate-headquarters/.git/modules/glazed')
newrepo = Path('/home/manuel/code/wesen/go-go-golems/glazed')

def git(cwd, args, check=True):
    p = subprocess.run(['git', '-C', str(cwd)] + args,
                       text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if check and p.returncode != 0:
        raise RuntimeError(p.stderr or p.stdout)
    return p

# Keep refs fresh.
git(newrepo, ['fetch', 'origin', '--prune'])
for remote in git(oldmeta, ['remote']).stdout.split():
    git(oldmeta, ['fetch', remote, '--prune'])

# Parse old worktrees.
items = []
cur = {}
for line in git(oldmeta, ['worktree', 'list', '--porcelain']).stdout.splitlines() + ['']:
    if not line:
        if cur:
            items.append(cur)
            cur = {}
        continue
    key, *rest = line.split(' ', 1)
    cur[key] = rest[0] if rest else ''

for item in items:
    wt = Path(item.get('worktree', ''))
    if not str(wt).startswith('/home/manuel/workspaces/'):
        continue

    branch_ref = item.get('branch', '')
    if not branch_ref.startswith('refs/heads/'):
        print('skip detached', wt)
        continue

    branch = branch_ref.removeprefix('refs/heads/')
    expected_head = item['HEAD']

    dirty = git(wt, ['status', '--porcelain=v1', '--untracked-files=all']).stdout.splitlines()
    unpushed = git(wt, ['rev-list', '--count', 'HEAD', '--not', '--remotes']).stdout.strip()
    current_head = git(wt, ['rev-parse', 'HEAD']).stdout.strip()

    if dirty or unpushed != '0' or current_head != expected_head:
        print('skip unsafe', wt, 'dirty=', len(dirty), 'unpushed=', unpushed)
        continue

    git(newrepo, ['fetch', str(oldmeta), f'+refs/heads/{branch}:refs/heads/{branch}'])
    copied = git(newrepo, ['rev-parse', branch]).stdout.strip()
    if copied != expected_head:
        raise RuntimeError(f'{branch}: copied {copied}, expected {expected_head}')

    git(oldmeta, ['worktree', 'remove', '--force', str(wt)])
    git(newrepo, ['worktree', 'add', str(wt), branch])

    if git(wt, ['rev-parse', '--verify', '--quiet', f'origin/{branch}'], check=False).returncode == 0:
        git(wt, ['branch', '--set-upstream-to', f'origin/{branch}', branch], check=False)

    assert git(wt, ['rev-parse', 'HEAD']).stdout.strip() == expected_head
```

## Handling dirty worktrees later

For each dirty worktree, choose one of these strategies.

### Commit WIP

```bash
cd "$wt"
git add -A
git commit -m 'WIP: preserve worktree changes before migration'
git push -u origin "$(git branch --show-current)"
```

Then migrate it as a clean worktree.

### Export patches instead of committing

Use this when the committed part of the worktree is already pushed/reachable, but the worktree has local uncommitted changes that should survive the migration without creating a WIP commit.

This path was tested on:

```text
/home/manuel/workspaces/2025-06-12/datadog-cli-tool/glazed
```

That worktree had one unstaged tracked modification, was exported, removed from the old `corporate-headquarters` common dir, recreated from `/home/manuel/code/wesen/go-go-golems/glazed`, and restored with an identical `git status --porcelain` result.

Export one dirty worktree:

```bash
wt=/home/manuel/workspaces/2025-06-12/datadog-cli-tool/glazed
repo=glazed
export_root=/tmp/worktree-migration

safe_name=$(echo "$wt" | sed 's#^/##; s#[/:]#_#g')
export_dir="$export_root/$repo/$safe_name"
mkdir -p "$export_dir"

# Metadata and before/after comparison input.
git -C "$wt" rev-parse HEAD > "$export_dir/HEAD"
git -C "$wt" branch --show-current > "$export_dir/branch"
git -C "$wt" status --porcelain=v1 --untracked-files=all > "$export_dir/status.before"

# Tracked changes. --binary is important for binary file changes.
git -C "$wt" diff --binary > "$export_dir/worktree.patch"
git -C "$wt" diff --cached --binary > "$export_dir/index.patch"

# Untracked, non-ignored files. Keep a NUL-separated manifest for inspection.
git -C "$wt" ls-files --others --exclude-standard -z > "$export_dir/untracked.list"
if [ -s "$export_dir/untracked.list" ]; then
  tar --null -C "$wt" -T "$export_dir/untracked.list" -czf "$export_dir/untracked.tgz"
else
  # Create a valid empty archive so restore can always run tar -xzf.
  tar -C "$wt" -czf "$export_dir/untracked.tgz" -T /dev/null
fi

echo "Exported dirty state to $export_dir"
```

Migrate the worktree while preserving its exact committed HEAD:

```bash
oldmeta=/home/manuel/code/wesen/corporate-headquarters/.git/modules/glazed
newrepo=/home/manuel/code/wesen/go-go-golems/glazed

branch=$(cat "$export_dir/branch")
expected_head=$(cat "$export_dir/HEAD")

# The committed part should be reachable from a remote before relying on patch export.
test "$(git -C "$wt" rev-list --count HEAD --not --remotes)" = 0

# Copy the local branch ref into the canonical clone before removing the old worktree.
git -C "$newrepo" fetch "$oldmeta" \
  "+refs/heads/$branch:refs/heads/$branch"
test "$(git -C "$newrepo" rev-parse "$branch")" = "$expected_head"

# Remove old worktree and recreate it from the canonical clone at the same path.
git -C "$oldmeta" worktree remove --force "$wt"
git -C "$newrepo" worktree add "$wt" "$branch"
test "$(git -C "$wt" rev-parse HEAD)" = "$expected_head"
```

Restore the exported dirty state:

```bash
# Apply only non-empty patch files; git apply fails on an empty file.
# Apply the index patch first, to both index and worktree, so staged adds/deletes
# exist in the working tree before the unstaged worktree patch is applied.
if [ -s "$export_dir/index.patch" ]; then
  git -C "$wt" apply --index "$export_dir/index.patch"
fi
if [ -s "$export_dir/worktree.patch" ]; then
  git -C "$wt" apply "$export_dir/worktree.patch"
fi

# Restore untracked files. This also works for the empty archive created above.
tar -C "$wt" -xzf "$export_dir/untracked.tgz"

# Compare restored status to the pre-migration status.
git -C "$wt" status --porcelain=v1 --untracked-files=all > "$export_dir/status.after"
diff -u "$export_dir/status.before" "$export_dir/status.after"

# Verify that the worktree now points at the canonical common dir.
sed -n '1p' "$wt/.git"
# expected: gitdir: /home/manuel/code/wesen/go-go-golems/glazed/.git/worktrees/...
```

Notes:

- This exports tracked unstaged changes, staged changes, and untracked non-ignored files.
- It does **not** export ignored files. Check `git -C "$wt" status --ignored --short` before removal if ignored files might matter.
- Recreate the worktree at the exact exported `HEAD`; applying patches after rebasing/updating the branch can create conflicts.
- Keep `$export_dir` until the restored status has been reviewed.
- If `git worktree remove --force "$wt"` removes the worktree metadata but fails to delete the directory because of permissions or generated caches, move the leftover directory aside, recreate the worktree at `$wt`, then restore from `$export_dir`. Record the leftover backup path so it can be deleted later after review.

## Post-migration checks

Count old pointers:

```bash
find /home/manuel/workspaces -name .git -type f -print |
while read -r gitfile; do
  if grep -q '/corporate-headquarters/.git/modules/glazed' "$gitfile"; then
    dirname "$gitfile"
  fi
done
```

Count new pointers:

```bash
find /home/manuel/workspaces -name .git -type f -print |
while read -r gitfile; do
  if grep -q '/go-go-golems/glazed/.git/worktrees' "$gitfile"; then
    dirname "$gitfile"
  fi
done
```

Check canonical worktrees:

```bash
git -C /home/manuel/code/wesen/go-go-golems/glazed worktree list
```

When all old worktrees have been migrated or intentionally deleted, prune the old common dir:

```bash
git -C "$oldmeta" worktree prune
```

Only remove the old common dir after there are no remaining `.git` files pointing to it.
