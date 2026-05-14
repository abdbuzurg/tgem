package helpers

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite testdata golden files instead of asserting")

// AssertJSONGolden normalizes `actual`, then either writes it to or compares
// it with golden/<name>.json. Pass -update-golden to rewrite. Diffs are
// reported line-by-line so test output is readable without an external diff
// tool.
func AssertJSONGolden(t *testing.T, name string, actual []byte) {
	t.Helper()

	normalized, err := Normalize(actual)
	if err != nil {
		t.Fatalf("normalize: %v\nraw: %s", err, string(actual))
	}
	assertGolden(t, name, normalized)
}

// AssertTextGolden compares `actual` against golden/<name> verbatim, no
// normalization. Pass -update-golden to rewrite. Use for non-JSON snapshots
// like the route-registry table.
func AssertTextGolden(t *testing.T, name string, actual []byte) {
	t.Helper()
	assertGolden(t, name, actual)
}

func assertGolden(t *testing.T, name string, content []byte) {
	t.Helper()

	path := goldenPath(name)
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for golden %s: %v", path, err)
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s (run with -update-golden to create): %v", path, err)
	}

	if string(want) == string(content) {
		return
	}

	t.Errorf("golden mismatch for %s\n%s", name, lineDiff(string(want), string(content)))
}

func goldenPath(name string) string {
	rel := filepath.FromSlash(name)
	if filepath.Ext(rel) == "" {
		rel += ".json"
	}
	return filepath.Join("test", "characterization", "golden", rel)
}

func lineDiff(want, got string) string {
	wl := strings.Split(want, "\n")
	gl := strings.Split(got, "\n")
	var b strings.Builder
	max := len(wl)
	if len(gl) > max {
		max = len(gl)
	}
	for i := 0; i < max; i++ {
		var w, g string
		if i < len(wl) {
			w = wl[i]
		}
		if i < len(gl) {
			g = gl[i]
		}
		if w == g {
			continue
		}
		fmt.Fprintf(&b, "  line %d:\n    want: %q\n    got:  %q\n", i+1, w, g)
	}
	return b.String()
}
