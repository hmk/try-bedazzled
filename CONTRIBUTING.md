# Contributing to try-bedazzled

Thanks for considering a contribution. This doc covers the essentials: how to
run the project locally, how we expect commits to be shaped, and how releases
happen.

## Local development

Requirements: Go 1.24+ (see `go.mod` for the exact version). A real terminal
is preferable for TUI work — most editors' integrated terminals work, but some
emoji/width quirks only show up in iTerm, Alacritty, WezTerm, etc.

```bash
# build
go build -o dist/try ./cmd/try

# run
./dist/try --version
./dist/try                       # opens the selector in an empty tries dir
./dist/try exec --path /tmp/try  # run against an arbitrary dir

# tests
go test ./...
go test ./... -race -count=1     # the version CI runs

# lint
go vet ./...
gofmt -l .                       # empty output = clean
staticcheck ./...                # install: `go install honnef.co/go/tools/cmd/staticcheck@latest`
```

For reproducible TUI testing against a populated tries dir, `bench/setup.sh`
generates 100 fixture directories under `/tmp/try-bench-fixtures`.

## Commit & PR style

We squash-merge PRs, so **the PR title becomes the commit message** on `main`.
PR titles must follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

feat: add a gradient selected-row background
fix: keep cursor on-screen when preview toggles
perf: reuse the fuzzy match buffer across rerenders
docs: clarify NO_COLOR behavior
```

Supported types: `feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `build`,
`ci`, `chore`, `revert`. A bot (`semantic-pr.yml`) validates titles on every
PR; it will fail the check if the title doesn't match.

Breaking changes: append `!` after the type (e.g. `feat!: rename config key`).
While pre-1.0 these still produce minor-version bumps.

## Release flow

We use [release-please](https://github.com/googleapis/release-please) to
automate versioning and changelog generation:

1. PRs merge to `main` with conventional-commit titles.
2. A `release-please` bot keeps a single open "release PR" updated with the
   next version number (based on the commits since the last release) and a
   generated `CHANGELOG.md` entry.
3. When the release PR is merged, release-please tags the commit (`v0.x.y`)
   and goreleaser publishes tarballs + deb/rpm to GitHub Releases.

Contributors don't need to touch version numbers or the changelog manually.

## Licensing

By submitting a pull request, you agree that your contribution is licensed
under the project's [BSD 3-Clause License](LICENSE). You retain copyright in
your contributions. We don't use a CLA.

If your PR includes third-party code, please note its license in the PR
description and make sure it's compatible with BSD-3-Clause.

## Code style

- Small, focused PRs are easier to review than sweeping ones.
- Prefer editing existing files over adding new ones.
- Don't add comments that restate what well-named code already says; comments
  are for explaining *why*, not *what*.
- Tests for user-visible behavior (TUI rendering, fuzzy scoring) are valued.
  Tests for internal plumbing are optional unless the plumbing is tricky.

## Getting help

File an issue with the bug or feature template. For security-sensitive
reports, see [SECURITY.md](SECURITY.md) instead.
