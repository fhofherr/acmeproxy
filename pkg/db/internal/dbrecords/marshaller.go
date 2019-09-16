package dbrecords

import (
	"bytes"
	"crypto"
	fmt "fmt"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
)

// MarshalBinary marshals the passed domain object into a byte slice suitable
// for storing it into the acmeproxy database.
func MarshalBinary(v interface{}) ([]byte, error) {
	m := &BinaryMarshaller{V: v}
	return m.MarshalBinary()
}

// BinaryMarshaller wraps a value V and provides a binary representation of value.
// Once the BinaryMarshaller was used to create a binary representation of V it
// must not be reused.
type BinaryMarshaller struct {
	V   interface{}
	err error
}

// MarshalBinary creates a binary representation of the object wrapped by
// the BinaryMarshaller.
func (m *BinaryMarshaller) MarshalBinary() ([]byte, error) {
	const op errors.Op = "dbrecords/binaryMarshaller.MarshalBinary"
	var bs []byte

	switch obj := m.V.(type) {
	case *acme.Client:
		bs = m.marshalACMEClient(obj)
	case acme.Client:
		bs = m.marshalACMEClient(&obj)
	case *acme.Domain:
		bs = m.marshalACMEDomain(obj)
	case acme.Domain:
		bs = m.marshalACMEDomain(&obj)
	case uuid.UUID:
		bs = m.marshalUUID(obj)
	case string:
		bs = []byte(obj)
	case *string:
		bs = []byte(*obj)
	default:
		return nil, errors.New(op, fmt.Sprintf("unsupported type: %T", m.V))
	}
	if m.err != nil {
		return bs, errors.New(op, fmt.Sprintf("marshal type: %T", m.V), m.err)
	}
	return bs, nil
}

func (m *BinaryMarshaller) marshalUUID(id uuid.UUID) []byte {
	const op errors.Op = "dbrecords/binaryMarshaller.marshalUUID"
	var (
		bs  []byte
		err error
	)
	m.do(func() error {
		bs, err = id.MarshalBinary()
		if err != nil {
			return errors.New(op, err)
		}
		return nil
	})
	return bs
}

func (m *BinaryMarshaller) marshalACMEClient(client *acme.Client) []byte {
	idBytes := m.marshalUUID(client.ID)
	kt, keyBytes := m.marshalPrivateKey(client.Key)
	rec := Client{
		Id:         idBytes,
		AccountURL: client.AccountURL,
		AccountKey: &Client_AccountKey{
			KeyType:  uint32(kt),
			KeyBytes: keyBytes,
		},
	}
	return m.marshalPB(&rec)
}

func (m *BinaryMarshaller) marshalPrivateKey(privateKey crypto.PrivateKey) (keyType, []byte) {
	const op errors.Op = "dbrecords/binaryMarshaller.marshalPrivateKey"

	var (
		certutilKt certutil.KeyType
		kt         keyType
		buf        bytes.Buffer
		err        error
	)
	m.do(func() error {
		certutilKt, err = certutil.DetermineKeyType(privateKey)
		if err != nil {
			return errors.New(op, err)
		}
		kt, err = keyTypeFromCertutil(certutilKt)
		if err != nil {
			return errors.New(op, err)
		}
		err = certutil.WritePrivateKey(privateKey, &buf, false)
		if err != nil {
			return errors.New(op, "write private key", err)
		}
		return nil
	})
	return kt, buf.Bytes()
}

func (m *BinaryMarshaller) marshalACMEDomain(d *acme.Domain) []byte {
	idBytes := m.marshalUUID(d.ClientID)
	rec := Domain{
		ClientID:       idBytes,
		Name:           d.Name,
		CertificatePEM: d.Certificate,
		PrivateKeyPEM:  d.PrivateKey,
	}
	return m.marshalPB(&rec)
}

func (m *BinaryMarshaller) marshalPB(msg proto.Message) []byte {
	const op errors.Op = "dbrecords/binaryMarshaller.marshalPB"
	var bs []byte
	m.do(func() error {
		var err error
		bs, err = proto.Marshal(msg)
		if err != nil {
			return errors.New(op, "marshal protobuf", err)
		}
		return nil
	})
	return bs
}

func (m *BinaryMarshaller) do(op func() error) {
	if m.err != nil {
		return
	}
	m.err = op()
}
