# Security policy

## Reporting a vulnerability

Please report vulnerabilities privately through GitHub:
**[Report a vulnerability](https://github.com/kinglet-dev/saymake/security/advisories/new)**
(Security tab → "Report a vulnerability").

Please don't open a public issue for security problems.

Include what you found, how to reproduce it (the description you typed, or
the file that was made, helps), and the impact you expect.

## What Saymake trusts

Nothing you type. Saymake runs entirely in your browser: the description and
the settings are read by the tool compiled to WebAssembly, within strict
limits (500 characters, 60 words, numbers of at most 4 digits and 2
decimals, settings clamped to their ranges, requests of at most 16 KB), and
anything outside them is refused. It has no server and sends nothing
anywhere: the single-file page's security policy forbids every network
connection, and on kinglet.dev the page may only load its own files from
kinglet.dev. Files are made only when you press a download button, from
numbers alone; no typed text ever reaches a file or a file name.

The full threat model is in [docs/DESIGN.md](docs/DESIGN.md#threat-model).

The machine-readable contact is published at
https://kinglet.dev/.well-known/security.txt (RFC 9116).
