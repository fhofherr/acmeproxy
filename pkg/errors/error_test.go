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
		{errors.Unauthorized, "unauthorized"},
		{errors.InvalidArgument, "invalid argument"},
		{errors.Kind(-1), "unknown kind"},
	}
	for _, tt := range tests {
		tt := tt
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
		tt := tt
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

func TestWrapError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		args     []interface{}
		expected error
	}{
		{
			name: "return nil if err is nil",
		},
		{
			name:     "wrap error if err is not nil",
			err:      errors.New("Oops"),
			args:     []interface{}{"something else"},
			expected: errors.New("something else", errors.New("Oops")),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := errors.Wrap(tt.err, tt.args...)
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.err.Error()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestError_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      *errors.Error
		target   error
		expected bool
	}{
		{
			name:   "target is not an errors.Error",
			err:    &errors.Error{},
			target: fmt.Errorf("some error"),
		},
		{
			name: "target is nil",
			err:  &errors.Error{},
		},
		{
			name:   "err is nil",
			target: &errors.Error{},
		},
		{
			name:     "both are nil",
			expected: true,
		},
		{
			name:     "err matches target",
			err:      &errors.Error{},
			target:   &errors.Error{},
			expected: true,
		},
		{
			name: "op does not match",
			err: &errors.Error{
				Op: "another op",
			},
			target: &errors.Error{
				Op: "some op",
			},
		},
		{
			name: "op matches",
			err: &errors.Error{
				Op:  "some op",
				Msg: "some message",
			},
			target: &errors.Error{
				Op: "some op",
			},
			expected: true,
		},
		{
			name: "kind does not match",
			err: &errors.Error{
				Kind: errors.Unspecified,
			},
			target: &errors.Error{
				Kind: errors.NotFound,
			},
		},
		{
			name: "kind matches",
			err: &errors.Error{
				Op:   "some op",
				Kind: errors.NotFound,
			},
			target: &errors.Error{
				Kind: errors.NotFound,
			},
			expected: true,
		},
		{
			name: "msg does not match",
			err: &errors.Error{
				Msg: "another message",
			},
			target: &errors.Error{
				Msg: "some message",
			},
		},
		{
			name: "msg matches",
			err: &errors.Error{
				Msg:  "some message",
				Kind: errors.NotFound,
			},
			target: &errors.Error{
				Msg: "some message",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Is(tt.target))
		})
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name     string
		err      *errors.Error
		target   error
		expected bool
	}{
		{
			name:   "target does not match",
			err:    &errors.Error{},
			target: fmt.Errorf("some error"),
		},
		{
			name:     "target matches exactly",
			err:      &errors.Error{},
			target:   &errors.Error{},
			expected: true,
		},
		{
			name: "nested error matches",
			err: &errors.Error{
				Msg: "another message",
				Err: &errors.Error{
					Msg: "some message",
				},
			},
			target: &errors.Error{
				Msg: "some message",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, errors.Is(tt.err, tt.target))
		})
	}
}

func TestGetKind(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected errors.Kind
	}{
		{
			name: "errors other than errors.Error are always unspecified",
			err:  fmt.Errorf("not one of our errors"),
		},
		{
			name:     "get kind of acmeproxy error",
			err:      errors.New(errors.NotFound),
			expected: errors.NotFound,
		},
		{
			name:     "the first specified Kind wins",
			err:      errors.New(errors.Unspecified, errors.New(errors.NotFound)),
			expected: errors.NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, errors.GetKind(tt.err))
			assert.True(t, errors.IsKind(tt.err, tt.expected))
		})
	}
}

func TestTrace(t *testing.T) {
	tests := []struct {
		name     string
		err      *errors.Error
		expected []errors.Op
	}{
		{
			name: "nil error",
		},
		{
			name: "error without op",
			err:  &errors.Error{},
			expected: []errors.Op{
				"unknown",
			},
		},
		{
			name: "error with op",
			err: &errors.Error{
				Op: errors.Op("op"),
			},
			expected: []errors.Op{
				errors.Op("op"),
			},
		},
		{
			name: "error with nested errors",
			err: &errors.Error{
				Op: errors.Op("op1"),
				Err: &errors.Error{
					Op: errors.Op("op2"),
				},
			},
			expected: []errors.Op{
				errors.Op("op1"),
				errors.Op("op2"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.err.Trace()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
