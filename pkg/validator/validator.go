package validator

import (
	"fmt"
	"strings"
)

// V collects field-level validation errors.
type V struct {
	errs map[string]string
}

// New creates a fresh validator.
func New() *V {
	return &V{errs: make(map[string]string)}
}

// Check adds an error for field if condition is false.
func (v *V) Check(condition bool, field, message string) {
	if !condition {
		v.errs[field] = message
	}
}

// Required fails if s is empty after trimming whitespace.
func (v *V) Required(s, field string) {
	v.Check(strings.TrimSpace(s) != "", field, fmt.Sprintf("%s is required", field))
}

// MaxLen fails if s exceeds n runes.
func (v *V) MaxLen(s string, n int, field string) {
	v.Check(len([]rune(s)) <= n, field, fmt.Sprintf("%s must be at most %d characters", field, n))
}

// MinLen fails if s is shorter than n runes.
func (v *V) MinLen(s string, n int, field string) {
	v.Check(len([]rune(s)) >= n, field, fmt.Sprintf("%s must be at least %d characters", field, n))
}

// OneOf fails if s is not in the allowed set.
func (v *V) OneOf(s string, allowed []string, field string) {
	for _, a := range allowed {
		if s == a {
			return
		}
	}
	v.Check(false, field, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")))
}

// Valid reports whether no errors have been recorded.
func (v *V) Valid() bool {
	return len(v.errs) == 0
}

// Errors returns the collected field errors map.
func (v *V) Errors() map[string]string {
	return v.errs
}
