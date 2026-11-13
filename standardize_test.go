package regexache

import (
	"strings"
	"testing"
)

func TestStandardize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected string
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
		},
		{
			TestName: "basic",
			Input:    `.*\S.*`,
			Expected: `.*\S.*`,
		},
		{
			TestName: "numOrder",
			Input:    `^[a-z0-9-_]+$`,
			Expected: `^[0-9a-z_-]+$`,
		},
		{
			TestName: "multiClass",
			Input:    `^[a-z0-9-_]+[a-z0-9-_]+$`,
			Expected: `^[0-9a-z_-]+[0-9a-z_-]+$`,
		},
		{
			TestName: "everywhere",
			Input:    `^[A-Za-z0-9-*&_]+$`,
			Expected: `^[\w-*&]+$`,
		},
		{
			TestName: "hex",
			Input:    `^[A-Fa-f0-9-*&_]+$`,
			Expected: `^[0-9A-Fa-f_-*&]+$`,
		},
		{
			TestName: "hex2",
			Input:    `^#[A-F0-9]{6}$`,
			Expected: `^#[0-9A-F]{6}$`,
		},
		{
			TestName: "parenthesis",
			Input:    `(/)|(/(([^~])|(~[01]))+)`,
			Expected: `(/)|(/(([^~])|(~[01]))+)`,
		},
		{
			TestName: "ordering",
			Input:    `^[a-zA-Z0-9]+$`,
			Expected: `^[0-9A-Za-z]+$`,
		},
		{
			TestName: "ordering2",
			Input:    `^[-a-zA-Z0-9._]*$`,
			Expected: `^[\w-.]*$`,
		},
		{
			TestName: "ordering3",
			Input:    `^[a-z0-9-]+$`,
			Expected: `^[0-9a-z-]+$`,
		},
		{
			TestName: "ordering4",
			Input:    `^[a-zA-Z0-9._-]*$`,
			Expected: `^[\w.-]*$`,
		},
		{
			TestName: "ordering5",
			Input:    `^[0-9a-zA-Z._-]+`,
			Expected: `^[\w.-]+`,
		},
		{
			TestName: "badEscaping",
			Input:    `[0-9a-zA-Z.\_\-]+$`,
			Expected: `[\w.\-]+$`,
		},
		{
			TestName: "badEscaping2",
			Input:    `[0-9a-zA-Z.\\_\-]+$`,
			Expected: `[0-9A-Za-z.\\_\-]+$`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got := standardize(testCase.Input)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

// TestPreloadIsStandardized guards preload.txt: entries already in standard
// form are hit by MustCompile whether or not REGEXACHE_STANDARDIZE is set.
func TestPreloadIsStandardized(t *testing.T) {
	t.Parallel()

	for r := range strings.SplitSeq(preload, "\n") {
		if r == "" {
			continue
		}
		if got := standardize(r); got != r {
			t.Errorf("preload.txt entry %q is not standardized; use %q", r, got)
		}
	}
}

func TestPreloadPatterns_Standardizing(t *testing.T) {
	originalStandardizing := standardizing
	defer func() { standardizing = originalStandardizing }()

	standardizing = true
	SetCaching(true)

	raw := `^[a-zA-Z0-9]+-preload-test$`
	preloadPatterns(raw)

	preloaded, ok := cache.Load(standardize(raw))
	if !ok {
		t.Fatalf("expected preload to store standardized key %q", standardize(raw))
	}
	if got := MustCompile(raw); got != preloaded {
		t.Error("expected MustCompile to return the preloaded instance")
	}
}
