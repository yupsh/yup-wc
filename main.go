// Command yup-wc is the CLI wrapper around github.com/gloo-foo/cmd-wc.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-wc"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name              = "wc"
	flagLines         = "lines"
	flagWords         = "words"
	flagBytes         = "bytes"
	flagChars         = "chars"
	flagMaxLineLength = "max-line-length"
)

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `wc [OPTIONS] [FILE...]

print newline, word, and byte counts for each FILE, and a total line if
more than one FILE is specified. A word is a non-zero-length sequence of
characters delimited by white space.
With no FILE, or when FILE is -, read standard input.`

// spec declares the wc wrapper: a file-or-stdin filter with wc's counting flags.
var spec = clix.Spec{
	Name:     name,
	Summary:  "print newline, word, and byte counts for each file",
	Synopsis: synopsis,
	Build:    build,
	Flags: []urf.Flag{
		&urf.BoolFlag{Name: flagLines, Aliases: []string{"l"}, Usage: "print the newline counts"},
		&urf.BoolFlag{Name: flagWords, Aliases: []string{"w"}, Usage: "print the word counts"},
		&urf.BoolFlag{Name: flagBytes, Aliases: []string{"c"}, Usage: "print the byte counts"},
		&urf.BoolFlag{Name: flagChars, Aliases: []string{"m"}, Usage: "print the character counts"},
		&urf.BoolFlag{Name: flagMaxLineLength, Aliases: []string{"L"}, Usage: "print the maximum display width"},
	},
}

// build maps the invocation to wc's pipeline: a file-or-stdin source into the wc
// command configured by the flags.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	return clix.OperandsOrStdin(inv), command.Wc(options(inv.Args)...), nil
}

// boolOption pairs a boolean flag's name with the Wc option it enables.
type boolOption struct {
	opt  any
	name string
}

// options folds the parsed flags into wc's option values.
func options(c *urf.Command) []any {
	flags := []boolOption{
		{name: flagLines, opt: command.WcLines},
		{name: flagWords, opt: command.WcWords},
		{name: flagBytes, opt: command.WcBytes},
		{name: flagChars, opt: command.WcChars},
		{name: flagMaxLineLength, opt: command.WcMaxLineLength},
	}
	var opts []any
	for _, f := range flags {
		if c.Bool(f.name) {
			opts = append(opts, f.opt)
		}
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
