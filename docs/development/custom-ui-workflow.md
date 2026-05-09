# Custom UI Development Workflow

This repository keeps official New API upgrade access while carrying local, lightweight frontend UI changes.

## Remote And Branch Roles

- `upstream`: official repository, `https://github.com/QuantumNous/new-api.git`.
- `origin`: personal fork, `git@github.com:2906189044/new-api-fork.git`.
- `main`: clean mirror of `upstream/main`. Do not commit local changes here.
- `custom/frontend-ui`: long-lived local customization branch. Put frontend UI changes here.
- `feature/ui-*`: short-lived branches for individual UI tasks, created from `custom/frontend-ui`.

Recommended local safety setting:

```bash
git remote set-url --push upstream DISABLED
```

This makes `upstream` effectively read-only and keeps all pushes pointed at `origin`.

## Daily UI Change Flow

Start each UI task from the customization branch:

```bash
git fetch --all --prune
git switch main
git reset --hard upstream/main
git switch custom/frontend-ui
git pull --rebase origin custom/frontend-ui
git rebase main
git push --force-with-lease origin custom/frontend-ui
git switch -c feature/ui-<short-name>
```

Keep changes scoped to `web/default/` unless the task explicitly needs backend support. Prefer additive custom files over editing upstream-heavy files. Good locations for low-conflict customization are:

- `web/default/src/styles/` for CSS variables, presets, and override files.
- `web/default/src/components/` for small wrapper components.
- `web/default/src/features/<feature>/` only when the page or feature itself must change.

After finishing a task:

```bash
cd web/default
bun install
bun run typecheck
bun run build
cd ../..
git switch custom/frontend-ui
git merge --ff-only feature/ui-<short-name>
git push origin custom/frontend-ui
```

Use normal merge instead of `--ff-only` only when the feature branch has intentionally diverged.

## Upgrade To Latest Official Version

Use this when official New API has new commits:

```bash
git fetch --all --prune
git switch main
git reset --hard upstream/main
git push --force-with-lease origin main:main

git switch custom/frontend-ui
git pull --rebase origin custom/frontend-ui
git rebase main
```

If conflicts occur, resolve them in favor of preserving local UI intent while keeping upstream behavior. Then verify:

```bash
cd web/default
bun install
bun run typecheck
bun run build
```

When verified:

```bash
cd ../..
git push --force-with-lease origin custom/frontend-ui
```

## Conflict Reduction Rules

- Do not develop on `main`.
- Do not modify protected upstream branding, attribution, package metadata, license headers, or project identity.
- Keep UI changes small and grouped by purpose.
- Prefer one commit per coherent UI change.
- Prefer additive CSS/custom components over broad rewrites of upstream layout files.
- Run frontend verification before pushing rebased customization branches.

## Optional Worktrees

For parallel UI tasks, create ignored local worktrees:

```bash
git worktree add .worktrees/ui-<short-name> -b feature/ui-<short-name> custom/frontend-ui
```

Remove a completed worktree after merging:

```bash
git worktree remove .worktrees/ui-<short-name>
```
