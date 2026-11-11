package apl

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"
)

const manualPath = "../Manual.md"

// Input lines of a transcript: after the 6-space prompt or a "[n] " prompt.
// A line of spaces and a caret is error output.
var (
	defPrompt = regexp.MustCompile(`^\[\d+\] `)
	caretLine = regexp.MustCompile(`^ *\^$`)
)

func transcriptInput(block string) string {
	var in []string
	for _, l := range strings.Split(block, "\n") {
		switch {
		case caretLine.MatchString(l):
		case strings.HasPrefix(l, "      "):
			in = append(in, l[6:])
		case l == "":
			// an empty output line
		case defPrompt.MatchString(l):
			in = append(in, defPrompt.ReplaceAllString(l, ""))
		}
	}
	return strings.Join(in, "\n") + "\n"
}

func transcript(input string) string {
	var out bytes.Buffer
	in := New(&out)
	in.Prompt = "      "
	in.Echo = true
	in.Run(strings.NewReader(input))
	return strings.TrimSuffix(out.String(), "\n")
}

// TestManual runs every ```apl block of Manual.md as a session and
// compares the transcript; -update rewrites the blocks.
func TestManual(t *testing.T) {
	data, err := os.ReadFile(manualPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	var result []string
	changed := false
	for i := 0; i < len(lines); i++ {
		result = append(result, lines[i])
		if lines[i] != "```apl" {
			continue
		}
		start := i + 1
		end := start
		for end < len(lines) && lines[end] != "```" {
			end++
		}
		block := strings.Join(lines[start:end], "\n")
		got := transcript(transcriptInput(block))
		if got != block {
			changed = true
			if !*update {
				t.Errorf("Manual.md:%d: example differs\n--- got\n%s\n--- want\n%s", start, got, block)
			}
		}
		result = append(result, strings.Split(got, "\n")...)
		i = end - 1
	}
	if *update && changed {
		if err := os.WriteFile(manualPath, []byte(strings.Join(result, "\n")), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
