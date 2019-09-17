package errors_test

import (
	"fmt"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestKindToString(t *testing.T) {
	tests := []struct {
		kind     errors.Kind
		expected string
	}{
		{errors.Unspecified, ""},
		{errors.NotFound, "not found"},
		{errors.Kind(-1), "unknown kind"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("kind %v: '%s'", tt.kind, tt.expected), func(t *testing.T) {
			actual := tt.kind.String()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected *errors.Error
	}{
		{
			name: "no args",
		},
		{
			name: "just a string",
			args: []interface{}{
				"something went wrong",
			},
			expected: &errors.Error{Msg: "something went wrong"},
		},
		{
			name: "op and error",
			args: []interface{}{
				errors.Op("some op"),
				fmt.Errorf("some error"),
			},
			expected: &errors.Error{
				Op:  "some op",
				Err: fmt.Errorf("some error"),
			},
		},
		{
			name: "op and a specific kind",
			args: []interface{}{
				errors.Op("some op"),
				errors.NotFound,
			},
			expected: &errors.Error{
				Op:   "some op",
				Kind: errors.NotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := errors.New(tt.args...)
			// If both expected and actual are nil we are ok. Otherwise
			// we check if expected and actual are equal.
			if tt.expected == nil && actual == nil {
				return
			}
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *errors.Error
		expected string
	}{
		{
			name: "nil error",
		},
		{
			name: "empty error",
			err:  &errors.Error{},
		},
		{
			name: "error message",
			err: &errors.Error{
				Msg: "something went wrong",
			},
			expected: "something went wrong",
		},
		{
			name: "op and wrapped arbitrary error",
			err: &errors.Error{
				Op:  errors.Op("test op"),
				Err: fmt.Errorf("something went wrong"),
			},
			expected: "test op: something went wrong",
		},
		{
			name: "error with kind and string",
			err: &errors.Error{
				Kind: errors.NotFound,
				Msg:  "something",
			},
			expected: "not found: something",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.err.Error()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestHasCause(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		cause    error
		expected bool
	}{
		{
			name:     "equal errors",
			err:      errors.New("Oops"),
			cause:    errors.New("Oops"),
			expected: true,
		},
		{
			name:     "error wraps cause",
			err:      errors.New("something else", errors.New("Oops")),
			cause:    errors.New("Oops"),
			expected: true,
		},
		{
			name:  "distinct errors",
			err:   errors.New("Oops"),
			cause: errors.New("something else"),
		},
		{
			name:  "cause not wrapped",
			err:   errors.New("something else", errors.New("not the droids you are looking for")),
			cause: errors.New("droids"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := errors.HasCause(tt.err, tt.cause)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
