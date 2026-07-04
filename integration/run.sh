#!/bin/sh
# Integration checks for yup-wc, run inside a Debian (GNU coreutils) container.
#
# yup-wc is NOT byte-compatible with GNU `wc`, so this harness uses `assert`
# (exact-output) checks rather than `parity` against the reference. Two
# documented divergences make GNU parity impossible (see cmd-wc COMPATIBILITY.md):
#   1. No field padding: yup-wc emits bare space-separated integers, while GNU
#      right-aligns each count in a common-width, leading-space-padded field and
#      (for file operands) appends the filename. With stdin GNU omits the name
#      but still pads.
#   2. Line terminators excluded: the gloo stream strips the trailing newline
#      from each line, so yup-wc's byte (-c) and char (-m) counts exclude the
#      newlines that GNU `wc` includes in its byte count.
#
# To make each divergence concrete, every assert also prints the GNU `wc`
# output for the same input so the difference is visible in the log.
set -eu

fails=0

# assert WANT FLAG INPUT — pipe INPUT into `yup-wc FLAG` and require WANT exactly.
# Also runs GNU `wc FLAG` on the same input for side-by-side comparison.
assert() {
	want=$1
	flag=$2
	input=$3
	got=$(printf '%s' "$input" | yup-wc $flag 2>/dev/null || true)
	gnu=$(printf '%s' "$input" | wc $flag 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  wc %-4s -> %-8s (gnu: %s)\n' "$flag" "$got" "$gnu"
	else
		printf 'FAIL  assert  wc %s\n        want: %s\n        got:  %s\n        gnu:  %s\n' "$flag" "$want" "$got" "$gnu"
		fails=$((fails + 1))
	fi
}

# Sample fixtures (trailing newline included so line counts are exact).
two='alpha
beta
'
words='one two three
four five
'
uni='abc
XY
'
long='hello world
foo
'

# Default: lines, words, bytes (bytes exclude the 2 stripped newlines: GNU = 11).
assert '2 2 9' '' "$two"

# -l: newline count (matches GNU's value, but GNU pads to a width).
assert '2' '-l' "$two"

# -w: word count (matches GNU's value; GNU pads).
assert '5' '-w' "$words"

# -c: byte count, EXCLUDING line terminators (GNU includes them -> 11).
assert '9' '-c' "$two"

# -m: character (rune) count, excluding newlines (GNU -m here -> 7 incl. newlines).
assert '5' '-m' "$uni"

# -L: longest line length in bytes (matches GNU's value; GNU pads).
assert '11' '-L' "$long"

# Multiple flags select multiple columns in GNU field order (lines words bytes).
assert '2 2 9' '-l -w -c' "$two"

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
