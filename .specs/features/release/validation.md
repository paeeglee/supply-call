# Release Validation

**Date**: 2026-09-27
**Spec**: `.specs/features/release/spec.md`
**Diff range**: `66e6972..HEAD` (a2ea27e..6f92547 plus 286da19; 3 commits on `ci/release-workflow`)
**Verifier**: independent sub-agent (author ≠ verifier)
**Verdict**: PASS

---

## Task Completion

No `tasks.md` (medium scope, inline plan). The 3 commits in range cover all spec surfaces:

| Commit | Scope | Status |
| ------ | ----- | ------ |
| 286da19 | `internal/release/version.go`, `version_test.go`, `cmd/nextversion/main.go` | ✅ Done |
| a2ea27e | `.github/workflows/release.yml` | ✅ Done |
| 6f92547 | `README.md` | ✅ Done |

---

## Spec-Anchored Acceptance Criteria

REL-01..07 are verified **by test** (`go test ./internal/release`). REL-08..15, 17..19 are workflow behavior that cannot run locally: verified **by inspection** of `.github/workflows/release.yml` (plus `actionlint` clean and local reproduction where possible). REL-16 by inspection of the README.

| AC | Spec-defined outcome | `file:line` + assertion / evidence | Method | Result |
| -- | -------------------- | ---------------------------------- | ------ | ------ |
| REL-01 | no `vX.Y.Z` tag → `v1.0.0` for any bump | `internal/release/version_test.go:12-15` cases (nil×patch/minor/major, only non-semver) → `"v1.0.0"`; assertion `version_test.go:27` `err != nil \|\| got != c.want`; impl `version.go:11,41-43` | test | ✅ PASS |
| REL-02 | `vA.B.C` + patch → `vA.B.(C+1)` | `version_test.go:16` `{"v1.2.3"}, "patch", "v1.2.4"`; impl `version.go:49-50` | test | ✅ PASS |
| REL-03 | `vA.B.C` + minor → `vA.(B+1).0` | `version_test.go:17` → `"v1.3.0"`; impl `version.go:47-48` | test | ✅ PASS |
| REL-04 | `vA.B.C` + major → `v(A+1).0.0` | `version_test.go:18` → `"v2.0.0"`; impl `version.go:45-46` | test | ✅ PASS |
| REL-05 | numeric max (`v1.10.0` > `v1.9.0`) | `version_test.go:19` → `"v1.10.1"`, `version_test.go:20` `v10.0.0` vs `v9.9.9` → `"v10.0.1"`; impl `version.go:37,55-62` | test | ✅ PASS |
| REL-06 | non-exact `vMAJOR.MINOR.PATCH` ignored (incl. `v1.0.0-beta`, leading zeros) | `version_test.go:21` prerelease → `"v1.0.1"`; `version_test.go:22` `v01.2.3`, `v1.2`, `v1.2.3.4`, `x1.5.0`, `" v3.0.0"` → `"v1.0.1"`; regex `version.go:13` | test | ✅ PASS |
| REL-07 | invalid bump → error; workflow stops before tag | `version_test.go:34-42` `err == nil` → `t.Errorf` for `""`,`Patch`,`build`,`1` with/without tags; impl `version.go:19-21`; CLI exits 1 via `cmd/nextversion/main.go:29-31` `log.Fatal` (reproduced: `bogus` → exit 1); workflow `.github/workflows/release.yml:50` `if ($LASTEXITCODE -ne 0) { exit 1 }` runs before tag step (`release.yml:76-84`). Input is also a `choice` (`release.yml:10-11`) | test + inspection | ✅ PASS |
| REL-08 | build with `build.ps1 <version without v>` → `main.version` | `release.yml:55` `number=$($tag.Substring(1))`; `release.yml:59` `./build.ps1 ${{ steps.version.outputs.number }}`; `build.ps1:26` `-X main.version=$Version` | inspection | ✅ PASS |
| REL-09 | generate/vet/test/build failure → no tag/Release | `build.ps1:4` `$ErrorActionPreference="Stop"`, `build.ps1:10` `throw` on non-zero exit for each step (`build.ps1:13,14,18/21,26`); failing Build step (`release.yml:58-59`) skips all later steps; tag/Release only created at `release.yml:83` | inspection | ✅ PASS |
| REL-10 | tag `vX.Y.Z` at `github.sha` + public Release with that name | `release.yml:83` `gh release create $env:TAG ... --target $env:GITHUB_SHA --title $env:TAG` (not `--draft`/`--prerelease`); `permissions: contents: write` `release.yml:14-15`. "Public" depends on repo visibility | inspection | ✅ PASS |
| REL-11 | notes contain full SHA + link | `release.yml:81` `$url = ".../commit/$env:GITHUB_SHA"`; `release.yml:82` `[$env:GITHUB_SHA]($url)` | inspection | ✅ PASS |
| REL-12 | Release asset `SupplyCall.exe` | `release.yml:83` positional asset `bin/SupplyCall.exe` (uploaded under basename `SupplyCall.exe`) | inspection | ✅ PASS |
| REL-13 | run artifact named `SupplyCall-vX.Y.Z` | `release.yml:72` `name: SupplyCall-${{ steps.version.outputs.tag }}`; `release.yml:73-74` path + `if-no-files-found: error` | inspection | ✅ PASS |
| REL-14 | non-`main` dispatch fails before build | `release.yml:29-33` first step `if: github.ref != 'refs/heads/main'` → `Write-Error` (terminating under runner's `$ErrorActionPreference='stop'`) / `exit 1` | inspection | ✅ PASS |
| REL-15 | concurrent dispatch queued, never parallel | `release.yml:18-20` `concurrency: group: release`, `cancel-in-progress: false` | inspection | ✅ PASS (see note N2) |
| REL-16 | README links `.../releases/latest/download/SupplyCall.exe` | `README.md:60` exact URL | inspection | ✅ PASS |
| REL-17 | tag already on remote → fail without overwriting | `release.yml:51-53` `git ls-remote --tags origin "refs/tags/$tag"`; non-empty → `Write-Error ... exit 1` before build/tag. Needed because `gh release create` on an existing tag without a Release would silently attach to it (N1) | inspection | ✅ PASS |
| REL-18 | tray "Versão X.Y.Z" == `-version` `Supply Call X.Y.Z` | single source: `main.go:33` `var version`; `-version` prints `main.go:98` `appName + " " + version`; tray gets `Version: version` `main.go:168` → `internal/tray/tray.go:101` `"Versão "+opts.Version`. Proven transitively by REL-19 check | inspection | ✅ PASS |
| REL-19 | `-version` ≠ `Supply Call X.Y.Z` → fail before tag/Release | `release.yml:65` `WANT: Supply Call ${{ steps.version.outputs.number }}`; `release.yml:66-67` `(& .\bin\SupplyCall.exe -version \| Out-String).Trim()` compared `-ne` → `exit 1`; runs before artifact/Release (`release.yml:70,76`). Reproduced locally: GUI-subsystem build (`-H windowsgui -X main.version=1.2.3`) piped via PowerShell yields `[Supply Call 1.2.3]`, exit 0 (stdout path `main.go:303`) | inspection + local repro | ✅ PASS |

**Status**: ✅ All 19 ACs covered (7 by test, 12 by inspection). No spec-precision gaps.

### Workflow pwsh semantics checked

- Runner wraps `shell: pwsh` with `$ErrorActionPreference = 'stop'` and an appended `exit $LASTEXITCODE`. `Write-Error` at `release.yml:32,53,67` therefore terminates the step itself; the following `exit 1` is unreachable but harmless.
- `$LASTEXITCODE` after the pipeline `git tag --list | go run ...` (`release.yml:49`) reflects the last native command (`go run`), so an invalid bump or a Go error stops the step (`release.yml:50`).
- `go run` build noise goes to stderr, so `$tag` holds only the single stdout line.
- `gh` failure is caught explicitly (`release.yml:84`).
- `actionlint` (latest, `-shellcheck=`) on `.github/workflows/release.yml`: exit 0, no findings. The pinned majors `actions/checkout@v7`, `actions/setup-go@v7`, `actions/upload-artifact@v7` exist upstream (`git ls-remote --tags`).
- `fetch-depth: 0` + `fetch-tags: true` (`release.yml:37-38`) guarantees `git tag --list` sees all tags.

---

## Discrimination Sensor

Isolated scratch: `git worktree add --detach <scratchpad>/wt HEAD`; mutations applied with `sed` to `internal/release/version.go` there, `go test -count=1 ./internal/release` run per mutant, worktree removed with `git worktree remove --force` + `prune`. Real tree `git status --porcelain` empty before and after (baseline matched).

| Mutation | File:line | Description | Killed? |
| -------- | --------- | ----------- | ------- |
| M1 | `internal/release/version.go:58` | `a[i] < b[i]` → `a[i] > b[i]` (picks lowest tag) | ✅ Killed |
| M2 | `internal/release/version.go:13` | regex tail accepts `-suffix` (prerelease counted) | ✅ Killed |
| M3 | `internal/release/version.go:48` | minor bump keeps patch | ✅ Killed |
| M4 | `internal/release/version.go:46` | major bump keeps minor | ✅ Killed |
| M5 | `internal/release/version.go:42` | first version `v0.0.1` instead of `First` | ✅ Killed |
| M6 | `internal/release/version.go:19` | bump validation removed (`false`) | ✅ Killed |
| M7 | `internal/release/version.go:50` | patch `+= 2` (off-by-one) | ✅ Killed |
| M8 | `internal/release/version.go:13` | major group `[0-9]+` (leading zeros accepted) | ✅ Killed |

**Sensor depth**: lightweight+ (8 mutations; the version math is the only unit-testable surface)
**Result**: 8/8 killed - PASS

The workflow YAML has no automated test; it cannot be mutation-tested locally.

---

## Code Quality

| Principle | Status |
| --------- | ------ |
| Minimum code | ✅ (`Next` + 1 helper, 31-line CLI, 84-line workflow) |
| Surgical changes | ✅ (no edits to `main.go`/`tray.go`/`build.ps1`, as spec assumes) |
| No scope creep | ✅ (no signing, changelog, multi-platform) |
| Matches patterns | ✅ (pure Go in `internal/`, thin `cmd/`, like AD-002) |
| Spec-anchored outcome check | ✅ (tests assert exact tag strings) |
| Per-layer coverage | ✅ domain 1:1 for REL-01..07; workflow by inspection only |
| Every test maps to a requirement | ✅ (the "blank lines ignored" case, `version_test.go:23`, supports the CLI reading `git tag --list` lines, an edge of REL-06) |
| Documented guidelines followed | none project-specific - strong defaults applied |

---

## Edge Cases

- [x] Calculated tag already on remote → fails before build, no overwrite (`release.yml:51-53`)
- [x] `v1.0.0` + `v1.0.0-beta` → only `v1.0.0` counts (`version_test.go:21`)

---

## Gate Check

- **Gate command**: `go test -count=1 ./internal/release` + `go vet ./internal/release ./cmd/nextversion` + `actionlint -shellcheck= .github/workflows/release.yml`
- **Result**: 2 tests passed (14 table cases + 8 invalid-bump combos), 0 failed, 0 skipped; vet clean; actionlint clean
- **Test count before feature**: 0 in `internal/release` (new package)
- **Test count after feature**: 2 test funcs
- **Delta**: +2
- **CLI smoke**: `v1.9.0,v1.10.0,foo` + minor → `v1.11.0`; empty + major → `v1.0.0`; `bogus` → exit 1

---

## Notes (non-blocking)

- **N1 (REL-17 robustness)**: the only guard against an existing *tag without a Release* is the `ls-remote` check (`release.yml:51-53`); `gh release create` alone would attach a Release to that existing tag (ignoring `--target`). The guard is present and correct; do not remove it.
- **N2 (REL-15)**: GitHub concurrency keeps at most one *pending* run per group; a third dispatch while one runs and one waits cancels the waiting one. Never parallel, as the spec requires, but not an unbounded queue.
- **N3 (REL-10/12/13/14/17/19)**: verified by inspection only; the real proof is the first dispatch on GitHub (the spec's Independent Test).

---

## Requirement Traceability Update

Recommended (spec.md not edited by the Verifier):

| Requirement | Previous Status | New Status |
| ----------- | --------------- | ---------- |
| REL-01..07 | Implementing | ✅ Verified (test) |
| REL-08..15, REL-17..19 | Implementing | ✅ Verified (inspection; confirm on first real run) |
| REL-16 | Implementing | ✅ Verified (inspection) |

---

## Summary

**Overall**: ✅ Ready

**Spec-anchored check**: 19/19 ACs matched the spec outcome (7 by test, 12 by inspection), 0 spec-precision gaps
**Sensor**: 8/8 mutations killed
**Gate**: 2 passed, 0 failed; actionlint clean

**Next steps**: dispatch the workflow on `main` after merge and check that `v1.0.0` exists with the `.exe`, that the tag SHA matches the notes, and that the artifact is `SupplyCall-v1.0.0`.
