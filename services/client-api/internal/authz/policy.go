package authz

import (
	"errors"
	"strings"
)

var ErrForbidden = errors.New("insufficient_scope")
var ErrFields = errors.New("invalid_fields")
var Summary = []string{"hazard_id", "source", "hazard_type", "severity", "area_name", "occurred_at", "ingested_at"}
var Raw = []string{"source_ref_id", "latitude", "longitude", "attributes"}

func Has(scope, want string) bool {
	for _, s := range strings.Fields(scope) {
		if s == want {
			return true
		}
	}
	return false
}
func Allowed(scope string, fields []string, raw bool) error {
	if !Has(scope, "hazard:read:summary") {
		return ErrForbidden
	}
	canRaw := Has(scope, "hazard:read:raw")
	if raw && !canRaw {
		return ErrForbidden
	}
	allowed := map[string]bool{}
	for _, s := range Summary {
		allowed[s] = true
	}
	for _, s := range Raw {
		allowed[s] = canRaw
	}
	// Check protected fields before unknown fields so explicit raw requests remain 403.
	for _, s := range fields {
		if ok, known := allowed[s]; known && !ok {
			return ErrForbidden
		}
	}
	for _, s := range fields {
		if _, known := allowed[s]; !known {
			return ErrFields
		}
	}
	return nil
}
