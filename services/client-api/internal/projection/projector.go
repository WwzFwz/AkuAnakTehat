package projection

import "example.com/akuanaktehat/client-api/internal/authz"

// A new object and explicit allowlist prevent future upstream fields leaking to Media.
func Project(h map[string]any, scope string, fields []string) map[string]any {
	allowed := append([]string{}, authz.Summary...)
	if authz.Has(scope, "hazard:read:raw") {
		allowed = append(allowed, authz.Raw...)
	}
	selected := map[string]bool{}
	for _, f := range fields {
		selected[f] = true
	}
	result := make(map[string]any)
	for _, f := range allowed {
		if len(fields) > 0 && !selected[f] {
			continue
		}
		if value, ok := h[f]; ok {
			result[f] = value
		}
	}
	return result
}
