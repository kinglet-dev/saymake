# Contributing

Thanks for your interest in Saymake.

## Before you start

- **Open an issue first** for anything beyond a typo fix, so we can agree on
  the change before you spend time on it. Descriptions the tool got wrong are
  especially welcome: say what you typed, what you expected and what it made.
- **Security problems:** don't open an issue; follow [SECURITY.md](SECURITY.md).

## Contributor License Agreement

Outside contributions can only be merged after the contributor signs a
Contributor License Agreement (CLA). The CLA process is not set up yet, so pull
requests from outside contributors can't be merged for now. Issues, bug reports
and descriptions it misreads are very welcome in the meantime.

## How changes are made

- Work on a branch (e.g. `feature/<short-name>`), never directly on `main`.
- Test first: write a failing test, then the code that makes it pass. Go
  code uses Ginkgo; coverage must stay at 100% (`scripts/check.sh` runs every
  check, including the page's unit and browser tests). `scripts/sweep.py`
  runs a mutation sweep over a package's conditions.
- Words are data: most vocabulary changes are edits to
  `adapters/vocabjson/default.json`, with an acceptance test in
  `acceptance/` for the description they fix.
- Keep each commit to one change with its tests; every commit builds and passes.
- Commit messages: a short imperative subject (72 characters at most), a blank
  line, then why the change was made.
- No new dependencies without discussion; any dependency must use a license
  that allows unrestricted commercial use (MIT, Apache 2.0, BSD, ISC).
- Don't commit secrets, personal data, local file paths, build output or
  coverage files.
