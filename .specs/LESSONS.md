# LESSONS - auto-maintained by scripts/lessons.py

> Machine-owned. Do NOT hand-edit. Changes are overwritten on the next `lessons.py` write.
> Canonical state lives in `.specs/lessons.json`. Edit lessons only via the script.
> promote_threshold=2 distinct features · window_days=45 · quarantine_threshold=2

## Confirmed (load these at Specify/Design)

Corroborated across multiple features. Safe to apply as guidance.

_none_

## Candidates (under observation - do NOT load as guidance yet)

Seen once or not yet corroborated. Tracked, not trusted.

### L-001 - For every numeric bound or sliding window named in an AC, add a test whose input crosses that bound so removing it changes the asserted value
- signal: `surviving_mutant` · recurrence: 1 feature(s) · scope: `clock` · harmful: 0
- features: overlay
- evidence: internal/clock/clock.go:80 (M2), internal/clock/clock.go:117 (M3), CLK-02 (clock)
- last seen: 2026-09-27T02:13:55Z

### L-002 - When one AC saves state on exit and another forbids overwriting a malformed file, save on exit only when the value actually changed
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `config` · harmful: 0
- features: overlay
- evidence: main.go:187 (CFG-03 vs TRAY-08) (config)
- last seen: 2026-09-27T02:13:55Z

### L-003 - Move save or skip decisions out of main into a pure function with a test covering the default or nil config case
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `main` · harmful: 0
- features: overlay
- evidence: main.go:190 (CFG-03 round 2, mutant M13 untestable) (main)
- last seen: 2026-09-27T02:19:50Z

## Quarantined (failed when applied - ignore)

A confirmed lesson that recurred alongside failure. Kept for the maintainer to review.

_none_
