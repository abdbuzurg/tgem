package helpers

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

var (
	timestampRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`)
)

// Normalize replaces non-deterministic values in a JSON document with stable
// sentinels so golden-file comparisons stay readable across runs.
//
// Rules (matching the plan):
//   - timestamp strings (RFC3339-ish)        -> "<TIMESTAMP>"
//   - field "id" or "*ID" with numeric > 0   -> "<ID>"  (key 0 is preserved
//     because the warehouse uses locationID=0)
//   - field "deliveryCode" (any string)      -> "<DELIVERY_CODE>"
//   - field "token" (any string)             -> "<TOKEN>"
//   - field "password" (any string)          -> "<PASSWORD_HASH>"  (bcrypt
//     output is non-deterministic)
func Normalize(in []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(in, &v); err != nil {
		return nil, err
	}
	v = walk(v, "")

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false) // keep "<TOKEN>" / "<ID>" sentinels readable
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func walk(v any, key string) any {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			x[k] = walk(child, k)
		}
		return x
	case []any:
		for i, child := range x {
			x[i] = walk(child, "")
		}
		return x
	case string:
		switch {
		case key == "token":
			return "<TOKEN>"
		case key == "password":
			return "<PASSWORD_HASH>"
		case key == "deliveryCode":
			return "<DELIVERY_CODE>"
		case timestampRE.MatchString(x):
			return "<TIMESTAMP>"
		}
		return x
	case float64:
		if isIDKey(key) && x > 0 {
			return "<ID>"
		}
		return x
	default:
		return v
	}
}

func isIDKey(key string) bool {
	if key == "id" {
		return true
	}
	// camelCase suffix "ID" — e.g. materialID, teamID, roleID, projectID. Keep
	// it case-sensitive because the codebase consistently uses camelCase.
	return len(key) > 2 && strings.HasSuffix(key, "ID")
}
