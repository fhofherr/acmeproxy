package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewMissingBucketFails(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	called := false
	err := fx.DB.viewBucket("missing_bucket", func(*bucket) error {
		called = true
		return nil
	})
	assert.Error(t, err)
	assert.False(t, called)
}

func TestViewErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		bucketError error
		viewError   error
	}{
		{
			name:        "handle bucket error",
			bucketError: errors.New("bucket error"),
		},
		{
			name:      "handle view function error",
			viewError: errors.New("view function error"),
		},
		{
			name:        "view function errors take precedence over bucket errors",
			bucketError: errors.New("bucket error"),
			viewError:   errors.New("view error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fx := NewTestFixture(t)
			defer fx.Close()

			bucketName := "some_bucket"
			fx.CreateBucket(bucketName)

			err := fx.DB.viewBucket(bucketName, func(b *bucket) error {
				b.Err = tt.bucketError
				return tt.viewError
			})
			assert.Error(t, err)
			if tt.viewError == nil {
				assert.Equal(t, tt.bucketError, err)
			} else {
				assert.Equal(t, tt.viewError, err)
			}
		})
	}
}

func TestCreateBucketBeforeUpdate(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	bucketName := "some_bucket"
	err := fx.DB.updateBucket(bucketName, func(b *bucket) error {
		assert.NotNil(t, b)
		return nil
	})
	assert.NoError(t, err)
	err = fx.DB.viewBucket(bucketName, func(b *bucket) error {
		assert.NotNil(t, b)
		return nil
	})
	assert.NoError(t, err)
}

func TestCreateBucketWithInvalidNameFails(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	called := false
	err := fx.DB.updateBucket("", func(*bucket) error {
		called = true
		return nil
	})
	assert.Error(t, err)
	assert.False(t, called)
}

func TestUpdateErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		bucketError error
		updateError error
	}{
		{
			name:        "handle bucket error",
			bucketError: errors.New("bucket error"),
		},
		{
			name:        "handle update function error",
			updateError: errors.New("view function error"),
		},
		{
			name:        "update function errors take precedence over bucket errors",
			bucketError: errors.New("bucket error"),
			updateError: errors.New("view error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fx := NewTestFixture(t)
			defer fx.Close()

			bucketName := "some_bucket"
			fx.CreateBucket(bucketName)

			err := fx.DB.updateBucket(bucketName, func(b *bucket) error {
				b.Err = tt.bucketError
				return tt.updateError
			})
			assert.Error(t, err)
			if tt.updateError == nil {
				assert.Equal(t, tt.bucketError, err)
			} else {
				assert.Equal(t, tt.updateError, err)
			}
		})
	}
}
