// Command goapl is an APL interpreter.
package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/sergev/goapl/apl"
	"github.com/sergev/goapl/internal/lineedit"
)

func main() {
	tty := isTerminal(os.Stdin)
	color := isTerminal(os.Stdout) && os.Getenv("NO_COLOR") == ""
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, tty, color))
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

const usage = "usage: goapl [-i] [file.apl ...]"

// run executes script files, or stdin if none; returns the exit status.
// After the files, ⎕LX runs if set; -i then continues with a session.
// tty enables the prompt and Ctrl-C to interrupt, color the ANSI colors.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer, tty, color bool) int {
	in := apl.New(stdout)
	in.Color = color
	session := len(args) == 0
	if len(args) > 0 && args[0] == "-i" {
		session, args = true, args[1:]
	}
	for _, name := range args {
		if strings.HasPrefix(name, "-") {
			fmt.Fprintln(stderr, usage)
			return 2
		}
		f, err := os.Open(name)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		in.Run(f)
		f.Close()
		if in.Off() {
			return 0
		}
	}
	if len(args) > 0 {
		in.RunLatent()
	}
	if !session {
		return 0
	}
	if tty {
		in.Prompt = "      "
		useEditor(in, stdin, stdout)
		intr := make(chan os.Signal, 1)
		signal.Notify(intr, os.Interrupt)
		defer signal.Stop(intr)
		go func() {
			for range intr {
				in.Interrupt()
			}
		}()
	}
	in.Run(stdin)
	return 0
}

// useEditor reads lines with the line editor when both stdin and stdout
// are terminals; history is kept in ~/.goapl_history.
func useEditor(in *apl.Interp, stdin io.Reader, stdout io.Writer) {
	fin, ok1 := stdin.(*os.File)
	fout, ok2 := stdout.(*os.File)
	if !ok1 || !ok2 || !isTerminal(fin) || !isTerminal(fout) {
		return
	}
	file := ""
	if home, err := os.UserHomeDir(); err == nil {
		file = filepath.Join(home, ".goapl_history")
	}
	ed := lineedit.New(fin, fout, file)
	in.ReadLine = func(prompt string) (string, error) {
		line, err := ed.ReadLine(prompt)
		if err == lineedit.ErrInterrupt {
			return "", nil
		}
		return line, err
	}
}
