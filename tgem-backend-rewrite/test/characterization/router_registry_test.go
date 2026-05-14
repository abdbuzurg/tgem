package characterization_test

import (
	"backend-v2/test/characterization/helpers"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// TestRouter_RouteRegistry snapshots every route registered by SetupRouter
// (method, path, fully-qualified handler symbol) into golden/router/routes.txt.
//
// During the Phase 3 layer rename this test is the safety net for the ~70%
// of routes the other characterization tests don't exercise. Each aggregate
// commit must change exactly the rows for the routes belonging to that
// aggregate — handler symbols flip from
//   backend-v2/internal/controller.(*fooController).Method-fm
// to
//   backend-v2/internal/http/handlers.(*fooHandler).Method-fm
// and nothing else. Any unrelated row changing or the route count drifting
// is itself the failure signal.
//
// Run `go test ./test/characterization -run TestRouter_RouteRegistry -update-golden`
// to rebase the golden after a deliberate change.
func TestRouter_RouteRegistry(t *testing.T) {
	routes := helpers.Engine().Routes()

	lines := make([]string, 0, len(routes))
	for _, r := range routes {
		lines = append(lines, fmt.Sprintf("%-7s %-70s %s", r.Method, r.Path, r.Handler))
	}
	sort.Strings(lines)

	actual := strings.Join(lines, "\n") + "\n"
	helpers.AssertTextGolden(t, "router/routes.txt", []byte(actual))
}
