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

// Marshal marshals the passed domain object into a byte slice suitable
// for storing it into the acmeproxy database.
func Marshal(v interface{}) ([]byte, error) {
	var (
		bs []byte
		m  = &marshaller{}
	)
	switch obj := v.(type) {
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
	default:
		return nil, errors.Errorf("unsupported type: %T", v)
	}
	return bs, errors.Wrapf(m.Err, "marshal type: %T", v)
}

type marshaller struct {
	Err error
}

func (m *marshaller) marshalUUID(id uuid.UUID) []byte {
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

func (m *marshaller) marshalACMEClient(client *acme.Client) []byte {
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

func (m *marshaller) marshalPrivateKey(privateKey crypto.PrivateKey) (keyType, []byte) {
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

func (m *marshaller) marshalACMEDomain(d *acme.Domain) []byte {
	idBytes := m.marshalUUID(d.ClientID)
	rec := Domain{
		ClientID:       idBytes,
		Name:           d.Name,
		CertificatePEM: d.Certificate,
		PrivateKeyPEM:  d.PrivateKey,
	}
	return m.marshalPB(&rec)
}

func (m *marshaller) marshalPB(msg proto.Message) []byte {
	var bs []byte
	m.do(func() error {
		var err error
		bs, err = proto.Marshal(msg)
		return errors.Wrap(err, "marshal protobuf")
	})
	return bs
}

func (m *marshaller) do(op func() error) {
	if m.Err != nil {
		return
	}
	m.Err = op()
}
