package db

import (
	"encoding"

	"github.com/fhofherr/acmeproxy/pkg/errors"
	"go.etcd.io/bbolt"
)

type bucket struct {
	Bkt *bbolt.Bucket
	Err error
}

func (b *bucket) writeRecord(id, record encoding.BinaryMarshaler) {
	const op errors.Op = "db/bucket.writeRecord"

	var (
		idBytes     []byte
		recordBytes []byte
	)
	if b.Err != nil {
		return
	}
	idBytes, b.Err = id.MarshalBinary()
	if b.Err != nil {
		return
	}
	recordBytes, b.Err = record.MarshalBinary()
	if b.Err != nil {
		return
	}
	b.Err = b.Bkt.Put(idBytes, recordBytes)
	if b.Err != nil {
		b.Err = errors.New(op, "put record", b.Err)
		return
	}
}

func (b *bucket) readRecord(id encoding.BinaryMarshaler, target encoding.BinaryUnmarshaler) {
	var (
		idBytes     []byte
		recordBytes []byte
	)
	if b.Err != nil {
		return
	}
	idBytes, b.Err = id.MarshalBinary()
	if b.Err != nil {
		return
	}
	recordBytes = b.Bkt.Get(idBytes)
	if recordBytes == nil {
		return
	}
	b.Err = target.UnmarshalBinary(recordBytes)
	if b.Err != nil {
		return
	}
}

// do executes f iff b.Err is nil. do makes it easy to write code that should
// only execute if the bucket has no error yet.
func (b *bucket) do(f func() error) {
	if b.Err != nil {
		return
	}
	b.Err = f()
}
