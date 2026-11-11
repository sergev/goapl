package apl

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/*.out")

// TestGolden runs testdata/*.apl and compares output with *.out. Each
// script runs in an empty directory, for workspace files.
func TestGolden(t *testing.T) {
	files, err := filepath.Glob("../testdata/*.apl")
	if err != nil {
		t.Fatal(err)
	}
	for i, f := range files {
		if files[i], err = filepath.Abs(f); err != nil {
			t.Fatal(err)
		}
	}
	for _, src := range files {
		t.Run(filepath.Base(src), func(t *testing.T) {
			f, err := os.Open(src)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			t.Chdir(t.TempDir())
			var out bytes.Buffer
			in := New(&out)
			in.Prompt = "      "
			in.Echo = true
			in.Run(f)
			want := strings.TrimSuffix(src, ".apl") + ".out"
			if *update {
				if err := os.WriteFile(want, out.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			exp, err := os.ReadFile(want)
			if err != nil {
				t.Fatal(err)
			}
			if got := out.String(); got != string(exp) {
				t.Errorf("output mismatch\n--- got\n%s--- want\n%s", got, exp)
			}
		})
	}
}
