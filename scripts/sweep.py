#!/usr/bin/env python3
"""Mutation sweep: for every `if` in the Go code (not tests, not cmd/), try
the negated condition, and for every comparison in it, the shifted boundary
(< ↔ <=, > ↔ >=). Each mutant is applied alone, its package's tests run with
a time limit, and the file is restored however the run ends (finally).
Mutants that don't compile are reported as INVALID, not as caught.
Usage: scripts/sweep.py [package dir ...]   (prints MISSED lines and a summary)"""
import os, re, subprocess, sys

IF = re.compile(r'^(\s*(?:\} else )?if )(.+?)( \{)$')
CMP = re.compile(r'(?<![<>=!])(<=|>=|<|>)(?![<>=])')

def mutants(line):
    m = IF.match(line)
    if not m:
        return
    head, cond, tail = m.groups()
    init = ""
    if "; " in cond:
        init, cond = cond.rsplit("; ", 1)
        init += "; "
    yield head + init + "!(" + cond + ")" + tail, "negated"
    for c in CMP.finditer(cond):
        op = {"<": "<=", "<=": "<", ">": ">=", ">=": ">"}[c.group(1)]
        yield head + init + cond[:c.start()] + op + cond[c.end():] + tail, f"{c.group(1)} -> {op}"

def run(pkg, args):
    return subprocess.run(["go", "test", "-count=1"] + args + ["./" + pkg], capture_output=True, timeout=180).returncode

def sweep(pkg):
    if run(pkg, []) != 0:
        print(f"BASELINE FAILS {pkg}"); return (0, 0, 0)
    caught = missed = invalid = 0
    for f in sorted(os.listdir(pkg)):
        if not f.endswith(".go") or f.endswith("_test.go"):
            continue
        path = os.path.join(pkg, f)
        orig = open(path).read()
        lines = orig.split("\n")
        try:
            for i, line in enumerate(lines):
                for mutated, what in mutants(line):
                    open(path, "w").write("\n".join(lines[:i] + [mutated] + lines[i + 1:]))
                    try:
                        if run(pkg, ["-run", "^$"]) != 0:
                            invalid += 1; continue
                        if run(pkg, []) == 0:
                            missed += 1; print(f"MISSED  {path}:{i + 1} {what}: {line.strip()}")
                        else:
                            caught += 1
                    except subprocess.TimeoutExpired:
                        caught += 1  # an endless loop is a failure the tests saw
        finally:
            open(path, "w").write(orig)
    print(f"{pkg}: {caught} caught, {missed} missed, {invalid} invalid", flush=True)
    return caught, missed, invalid

if __name__ == "__main__":
    pkgs = sys.argv[1:] or sorted({os.path.dirname(os.path.join(r, f))[2:] for r, _, fs in os.walk(".")
        for f in fs if f.endswith(".go") and not f.endswith("_test.go") and "/cmd/" not in r and "node_modules" not in r})
    totals = [sum(t) for t in zip(*(sweep(p) for p in pkgs))]
    print(f"TOTAL: {totals[0]} caught, {totals[1]} missed, {totals[2]} invalid")
