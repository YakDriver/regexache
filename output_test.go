package regexache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputCache_OneFilePerProcess(t *testing.T) {
	originalOutputFile := outputFile
	originalLookups := lookups
	defer func() {
		outputFile = originalOutputFile
		lookups = originalLookups
	}()

	dir := t.TempDir()
	outputFile = filepath.Join(dir, "out")
	lookups = make(map[string]int)
	SetCaching(true)

	MustCompile("output-one")
	outputCache()
	MustCompile("output-one")
	MustCompile("output-two")
	outputCache()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("out.%d", os.Getpid())
	if len(entries) != 1 || entries[0].Name() != want {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("expected only %s, got %v", want, names)
	}

	data, err := os.ReadFile(filepath.Join(dir, want))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"regex\tcount\n", "output-one\t2\n", "output-two\t1\n"} {
		if !strings.Contains(string(data), s) {
			t.Errorf("expected latest snapshot to contain %q, got:\n%s", s, data)
		}
	}

	// Output must get the same permissions as a plain 0644 create under the
	// current umask, not os.CreateTemp's owner-only 0600.
	ref := filepath.Join(t.TempDir(), "ref")
	if err := os.WriteFile(ref, nil, 0644); err != nil {
		t.Fatal(err)
	}
	refInfo, err := os.Stat(ref)
	if err != nil {
		t.Fatal(err)
	}
	outInfo, err := os.Stat(filepath.Join(dir, want))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := outInfo.Mode().Perm(), refInfo.Mode().Perm(); got != want {
		t.Errorf("output file mode = %v, want %v", got, want)
	}
}
