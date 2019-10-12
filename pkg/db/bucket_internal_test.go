package db

import (
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/db/internal/dbrecords"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWriteAndReadRecord(t *testing.T) {
	var (
		actual     string
		id         = uuid.Must(uuid.NewRandom())
		record     = "some record"
		bucketName = "some_bucket"
	)

	fx := NewTestFixture(t)
	defer fx.Close()

	err := fx.DB.updateBucket(bucketName, func(b *bucket) error {
		b.writeRecord(id, &dbrecords.BinaryMarshaller{V: record})
		return b.Err
	})
	if !assert.NoError(t, err) {
		return
	}
	err = fx.DB.viewBucket(bucketName, func(b *bucket) error {
		target := &dbrecords.BinaryUnmarshaller{V: &actual}
		b.readRecord(id, target)
		return b.Err
	})

	assert.NoError(t, err)
	assert.Equal(t, record, actual)
}

func TestReadRecordCantMarshalID(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	marshalErr := errors.New("can't marshal")
	id := newMockBinaryMarshaller(t)
	id.On("MarshalBinary").Return([]byte(nil), marshalErr)
	target := newMockBinaryUnmarshaller(t)

	bucketName := "some_bucket"
	fx.CreateBucket(bucketName)
	err := fx.DB.viewBucket(bucketName, func(b *bucket) error {
		b.readRecord(id, target)
		return b.Err
	})

	assert.Error(t, err)
	assert.Equal(t, marshalErr, err)
	id.AssertCalled(t, "MarshalBinary")
	target.AssertNotCalled(t, "UnmarshalBinary", mock.Anything)
}

func TestReadRecordCantUnmarshalRecord(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	record := "some record"
	unmarshalErr := errors.New("can't unmarshal")
	id := uuid.Must(uuid.NewRandom())
	target := newMockBinaryUnmarshaller(t)
	target.On("UnmarshalBinary", []byte(record)).Return(unmarshalErr)

	bucketName := "some_bucket"
	fx.CreateBucketWithKey(bucketName, id, &dbrecords.BinaryMarshaller{V: record})

	err := fx.DB.viewBucket(bucketName, func(b *bucket) error {
		b.readRecord(id, target)
		return b.Err
	})
	assert.Error(t, err)
	assert.Error(t, unmarshalErr, err)
	target.AssertCalled(t, "UnmarshalBinary", []byte(record))
}

func TestReadRecordDoesNothingIfBucketHasError(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	bucketErr := errors.New("bucket error")
	id := newMockBinaryMarshaller(t)
	record := newMockBinaryUnmarshaller(t)

	bucketName := "some_bucket"
	fx.CreateBucket(bucketName)
	err := fx.DB.viewBucket(bucketName, func(b *bucket) error {
		b.Err = bucketErr
		b.readRecord(id, record)
		return b.Err
	})

	assert.Error(t, err)
	assert.Equal(t, bucketErr, err)
}

func TestWriteRecordCantMarshalID(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	marshalErr := errors.New("can't marshal id")
	id := newMockBinaryMarshaller(t)
	id.On("MarshalBinary").Return([]byte(nil), marshalErr)
	record := newMockBinaryMarshaller(t)

	err := fx.DB.updateBucket("some_bucket", func(b *bucket) error {
		b.writeRecord(id, record)
		return b.Err
	})
	assert.Error(t, err)
	assert.Equal(t, marshalErr, err)
	id.AssertCalled(t, "MarshalBinary")
	record.AssertNotCalled(t, "MarshalBinary")
}

func TestWriteRecordCantMarshalRecord(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	id := uuid.Must(uuid.NewRandom())
	marshalErr := errors.New("can't marshal record")
	record := newMockBinaryMarshaller(t)
	record.On("MarshalBinary").Return([]byte(nil), marshalErr)

	err := fx.DB.updateBucket("some_bucket", func(b *bucket) error {
		b.writeRecord(id, record)
		return b.Err
	})
	assert.Error(t, err)
	assert.Equal(t, marshalErr, err)
	record.AssertCalled(t, "MarshalBinary")
}

func TestWriteRecordPutToBucketFails(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	record := "some record"
	err := fx.DB.updateBucket("some_bucket", func(b *bucket) error {
		b.writeRecord(
			// An empty id makes bolt fail, as the id is used as key which may
			// not be empty.
			&dbrecords.BinaryMarshaller{V: ""},
			&dbrecords.BinaryMarshaller{V: record},
		)
		return b.Err
	})
	assert.Error(t, err)
}

func TestWriteRecordDoesNothingIfBucketHasError(t *testing.T) {
	fx := NewTestFixture(t)
	defer fx.Close()

	bucketErr := errors.New("bucket error")
	id := newMockBinaryMarshaller(t)
	record := newMockBinaryMarshaller(t)
	err := fx.DB.updateBucket("some_bucket", func(b *bucket) error {
		b.Err = bucketErr
		b.writeRecord(id, record)
		return b.Err
	})

	assert.Error(t, err)
	assert.Equal(t, bucketErr, err)
}

func TestDoExecutesCodeIffBucketIsOk(t *testing.T) {
	tests := []struct {
		name        string
		bucketError error
		funcError   error
		funcCalled  bool
	}{
		{
			name:       "bucket is ok",
			funcCalled: true,
		},
		{
			name:        "bucket has error",
			bucketError: errors.New("bucket error"),
			funcCalled:  false,
		},
		{
			name:       "function fails bucket",
			funcError:  errors.New("function error"),
			funcCalled: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fx := NewTestFixture(t)
			defer fx.Close()

			bucketName := "some_bucket"
			fx.CreateBucket(bucketName)

			called := false
			f := func() error {
				called = true
				return tt.funcError
			}
			err := fx.DB.viewBucket(bucketName, func(b *bucket) error {
				b.Err = tt.bucketError
				b.do(f)
				return nil
			})
			if tt.funcError != nil {
				assert.Equal(t, tt.funcError, err)
			} else {
				assert.Equal(t, tt.bucketError, err)
			}
			assert.Equal(t, tt.funcCalled, called)
		})
	}
}

type mockBinaryMarshaller struct {
	mock.Mock
}

func newMockBinaryMarshaller(t *testing.T) *mockBinaryMarshaller {
	m := &mockBinaryMarshaller{}
	m.Test(t)
	return m
}

func (m *mockBinaryMarshaller) MarshalBinary() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

type mockBinaryUnmarshaller struct {
	mock.Mock
}

func newMockBinaryUnmarshaller(t *testing.T) *mockBinaryUnmarshaller {
	m := &mockBinaryUnmarshaller{}
	m.Test(t)
	return m
}

func (m *mockBinaryUnmarshaller) UnmarshalBinary(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}
