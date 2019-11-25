package errors

import "testing"

// AssertMatches checks if the actual error matches the passed template error.
func AssertMatches(t *testing.T, template, actual error) bool {
	if !Is(actual, template) {
		t.Errorf("'%v' did not match template '%v'", actual, template)
		return false
	}
	return true
}
