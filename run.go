package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-wc"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "wc"

const (
	flagLines         = "lines"
	flagWords         = "words"
	flagBytes         = "bytes"
	flagChars         = "chars"
	flagMaxLineLength = "max-line-length"
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `wc [OPTIONS] [FILE...]

print newline, word, and byte counts for each FILE, and a total line if
more than one FILE is specified. A word is a non-zero-length sequence of
characters delimited by white space.
With no FILE, or when FILE is -, read standard input.`

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the wc CLI against the injected version, I/O, and
// filesystem, returning the process exit code.
func run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newCommand(version, stdin, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newCommand(version string, stdin io.Reader, stdout io.Writer, fs afero.Fs) *cli.Command {
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "print newline, word, and byte counts for each file",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: flagLines, Aliases: []string{"l"}, Usage: "print the newline counts"},
			&cli.BoolFlag{Name: flagWords, Aliases: []string{"w"}, Usage: "print the word counts"},
			&cli.BoolFlag{Name: flagBytes, Aliases: []string{"c"}, Usage: "print the byte counts"},
			&cli.BoolFlag{Name: flagChars, Aliases: []string{"m"}, Usage: "print the character counts"},
			&cli.BoolFlag{Name: flagMaxLineLength, Aliases: []string{"L"}, Usage: "print the maximum display width"},
		},
		Action: action(stdin, stdout, fs),
	}
}

func action(stdin io.Reader, stdout io.Writer, fs afero.Fs) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		_, err := gloo.Run(source(c, stdin, fs), gloo.ByteWriteTo(stdout), command.Wc(options(c)...))
		return err
	}
}

func source(c *cli.Command, stdin io.Reader, fs afero.Fs) any {
	if c.NArg() == 0 {
		return gloo.ByteReaderSource([]io.Reader{stdin})
	}
	files := make([]gloo.File, c.NArg())
	for i := range files {
		files[i] = gloo.File(c.Args().Get(i))
	}
	return gloo.ByteFileSource(fs, files)
}

func options(c *cli.Command) []any {
	flags := []struct {
		opt  any
		name string
	}{
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
