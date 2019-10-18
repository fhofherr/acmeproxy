package errors_test

import (
	"fmt"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCollection_Append(t *testing.T) {
	tests := []struct {
		name     string
		initial  errors.Collection
		err      error
		args     []interface{}
		expected errors.Collection
	}{
		{
			name: "nil collection, nil error",
		},
		{
			name: "nil collection, non-nil error",
			err:  fmt.Errorf("some error"),
			args: []interface{}{
				errors.Op("some op"),
			},
			expected: errors.Collection{
				errors.Wrap(fmt.Errorf("some error"), errors.Op("some op")),
			},
		},
		{
			name: "non-nil collection, nil error",
			initial: errors.Collection{
				errors.New("some error"),
			},
			args: []interface{}{
				errors.Op("some op"),
			},
			expected: errors.Collection{
				errors.New("some error"),
			},
		},

		{
			name: "non-nil collection, non-nil error",
			initial: errors.Collection{
				errors.New("some error"),
			},
			err: fmt.Errorf("some other error"),
			args: []interface{}{
				errors.Op("some op"),
			},
			expected: errors.Collection{
				errors.New("some error"),
				errors.Wrap(fmt.Errorf("some other error"), errors.Op("some op")),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := errors.Append(tt.initial, tt.err, tt.args...)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestCollection_Error(t *testing.T) {
	tests := []struct {
		name       string
		collection errors.Collection
		expected   string
	}{
		{
			name: "nil collection",
		},
		{
			name:       "empty collection",
			collection: errors.Collection{},
		},
		{
			name: "single entry",
			collection: errors.Collection{
				errors.New("single error"),
			},
			expected: errors.New("single error").Error(),
		},
		{
			name: "multiple entries",
			collection: errors.Collection{
				errors.New("first error"),
				errors.New("second error"),
			},
			expected: fmt.Sprintf(
				"* %s\n* %s",
				errors.New("first error").Error(),
				errors.New("second error").Error(),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, tt.collection, tt.expected)
		})
	}
}

func TestCollection_ErrorOrNil(t *testing.T) {
	tests := []struct {
		name       string
		collection errors.Collection
		expected   error
	}{
		{
			name: "nil collection",
		},
		{
			name:       "empty collection",
			collection: errors.Collection{},
		},
		{
			name: "non-empty collection",
			collection: errors.Collection{
				errors.New("some error"),
			},
			expected: errors.Collection{
				errors.New("some error"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.collection.ErrorOrNil())
		})
	}
}
