package dbrecords

import (
	"bytes"
	"crypto"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
		return nil, errors.Errorf("unsupported type: %T", m.V)
	}
	return bs, errors.Wrapf(m.err, "marshal type: %T", m.V)

}

func (m *BinaryMarshaller) marshalUUID(id uuid.UUID) []byte {
	var (
		bs  []byte
		err error
	)
	m.do(func() error {
		bs, err = id.MarshalBinary()
		return errors.Wrap(err, "marshall id bytes")
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
	var (
		certutilKt certutil.KeyType
		kt         keyType
		buf        bytes.Buffer
		err        error
	)
	m.do(func() error {
		certutilKt, err = certutil.DetermineKeyType(privateKey)
		if err != nil {
			return errors.Wrap(err, "determine key type")
		}
		kt, err = keyTypeFromCertutil(certutilKt)
		if err != nil {
			return errors.Wrap(err, "convert certutil key type")
		}
		err = certutil.WritePrivateKey(privateKey, &buf, false)
		return errors.Wrap(err, "write private key")
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
	var bs []byte
	m.do(func() error {
		var err error
		bs, err = proto.Marshal(msg)
		return errors.Wrap(err, "marshal protobuf")
	})
	return bs
}

func (m *BinaryMarshaller) do(op func() error) {
	if m.err != nil {
		return
	}
	m.err = op()
}
