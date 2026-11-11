#!/usr/bin/env python3
"""Compare goapl with the Pascal prototype (pnscanner) on a case file.

Usage: compare.py PROTO_DIR GOAPL CASES

PROTO_DIR holds pnscanner.pas and APLfile.txt; the prototype is built
with fpc into build/proto. Each case line holds statements separated
by " ⋄ ". A line starting with "!" is a known difference: it is run
and shown, but does not count as a failure.
"""
import os
import re
import shutil
import subprocess
import sys

# APLCharSet order, matching the first line of APLfile.txt.
CHARSET = ("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890:→←∘.()[];⎕ +-×÷*⍳⍴,~=≠<≤≥>∧∨⌈⌊○/\"¯?ωε↑↓α_∇∆'⊃⊂∩∪⊥⊤|\\")
IGNORED = {"NOT USED", "INPUT", "NIL", "RE-ENTER LAST LINE"}


def build(proto_dir):
    out = os.path.join("build", "proto")
    os.makedirs(out, exist_ok=True)
    for f in ("pnscanner.pas", "APLfile.txt"):
        shutil.copy(os.path.join(proto_dir, f), out)
    exe = os.path.join(out, "pnscanner")
    src = os.path.join(out, "pnscanner.pas")
    if not os.path.exists(exe) or os.path.getmtime(exe) < os.path.getmtime(src):
        subprocess.run(["fpc", "-Miso", "-v0", "pnscanner.pas"], cwd=out, check=True,
                       stdout=subprocess.DEVNULL)
    with open(os.path.join(out, "APLfile.txt"), "rb") as f:
        lines = f.read().split(b"\n")
    bytemap = dict(zip(CHARSET, lines[0]))
    messages = {l.decode("latin-1").strip() for l in lines[1:101]} - IGNORED
    return out, bytemap, messages


def encode(text, bytemap):
    return bytes(bytemap[c] if c in bytemap else ord(c) for c in text)


def found(out, messages):
    out = out.upper()
    return sorted(m for m in messages if m in out)


def run_proto(lines, dir, bytemap, messages):
    data = b"".join(encode(l, bytemap) + b"\n" for l in lines) + b"/*\n"
    p = subprocess.run(["./pnscanner", "APLfile.txt"], input=data, cwd=dir,
                       capture_output=True, timeout=5)
    out = p.stdout.decode("latin-1")
    if p.returncode != 0:
        return [], ["CRASH"], out
    neg = chr(bytemap["¯"])
    text = re.sub(re.escape(neg) + r"\s*", "-", out)
    return [float(x) for x in re.findall(r"-?\d+\.\d+", text)], found(out, messages), out


def run_goapl(lines, exe, messages):
    p = subprocess.run([exe], input=("\n".join(lines) + "\n").encode(),
                       capture_output=True, timeout=5)
    out = p.stdout.decode()
    nums = re.findall(r"¯?\d+(?:\.\d+)?(?:E¯?\d+)?", out)
    return [float(x.replace("¯", "-").replace("E", "e")) for x in nums], found(out, messages), out


def agree(p, g):
    pn, pm, _ = p
    gn, gm, _ = g
    if pm or gm:
        return pm == gm
    return len(pn) == len(gn) and all(abs(a - b) <= 0.006 + 1e-9 * abs(a) for a, b in zip(pn, gn))


def main():
    proto_dir, exe, cases_file = sys.argv[1:4]
    dir, bytemap, messages = build(proto_dir)
    exe = os.path.abspath(exe)
    fails = known = total = 0
    for raw in open(cases_file, encoding="utf-8"):
        raw = raw.rstrip("\n")
        if not raw.strip() or raw.startswith("#"):
            continue
        expect_diff = raw.startswith("!")
        lines = raw.lstrip("!").split(" ⋄ ")
        total += 1
        p = run_proto(lines, dir, bytemap, messages)
        g = run_goapl(lines, exe, messages)
        ok = agree(p, g)
        if ok and not expect_diff:
            continue
        if expect_diff and not ok:
            known += 1
            tag = "KNOWN"
        else:
            fails += 1
            tag = "FAIL" if not ok else "FAIL (marked ! but agrees)"
        print(f"{tag}: {' ⋄ '.join(lines)}")
        print(f"  proto: {p[1] or p[0]}")
        print(f"  goapl: {g[1] or g[0]}")
    print(f"{total - fails - known} agree, {known} known differences, {fails} failures")
    return 1 if fails else 0


if __name__ == "__main__":
    sys.exit(main())
