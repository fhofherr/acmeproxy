package dbrecords

import (
	"crypto"
	"crypto/x509"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Unmarshal unmarshals the passed byte slice into v which should be a pointer
// to a passed domain object.
func Unmarshal(bs []byte, v interface{}) error {
	u := &unmarshaller{}
	switch obj := v.(type) {
	case *acme.Client:
		u.unmarshalACMEClient(bs, obj)
	case *acme.Domain:
		u.unmarshalACMEDomain(bs, obj)
	case *uuid.UUID:
		u.unmarshalUUID(bs, obj)
	default:
		return errors.Errorf("unsupported type: %T", v)
	}
	return errors.Wrapf(u.Err, "unmarshal type: %T", v)
}

type unmarshaller struct {
	Err error
}

func (u *unmarshaller) unmarshalUUID(bs []byte, id *uuid.UUID) {
	u.do(func() error {
		err := id.UnmarshalBinary(bs)
		return errors.Wrap(err, "unmarshal uuid")
	})
}

func (u *unmarshaller) unmarshalACMEClient(bs []byte, client *acme.Client) {
	u.do(func() error {
		if client == nil {
			return errors.New("client must not be nil")
		}
		var (
			rec Client
			id  uuid.UUID
		)
		u.unmarshalPB(bs, &rec)
		u.unmarshalUUID(rec.Id, &id)
		key := u.unmarshalPrivateKey(keyType(rec.AccountKey.KeyType), rec.AccountKey.KeyBytes)
		client.AccountURL = rec.AccountURL
		client.ID = id
		client.Key = key
		return nil
	})
}

func (u *unmarshaller) unmarshalPrivateKey(kt keyType, bs []byte) crypto.PrivateKey {
	var pk crypto.PrivateKey
	u.do(func() error {
		var err error
		switch kt {
		case ecdsa:
			pk, err = x509.ParseECPrivateKey(bs)
			return errors.Wrap(err, "unmarshal ECDSA private key")
		case rsa:
			pk, err = x509.ParsePKCS1PrivateKey(bs)
			return errors.Wrap(err, "parse RSA private key")
		default:
			return errors.Errorf("unknown key type: %v", kt)
		}
	})
	return pk
}

func (u *unmarshaller) unmarshalACMEDomain(bs []byte, domain *acme.Domain) {
	u.do(func() error {
		if domain == nil {
			return errors.New("domain must not be nil")
		}
		var (
			rec      Domain
			clientID uuid.UUID
		)
		u.unmarshalPB(bs, &rec)
		u.unmarshalUUID(rec.ClientID, &clientID)
		domain.ClientID = clientID
		domain.Name = rec.Name
		domain.Certificate = rec.CertificatePEM
		domain.PrivateKey = rec.PrivateKeyPEM
		return nil
	})
}

func (u *unmarshaller) unmarshalPB(bs []byte, msg proto.Message) {
	u.do(func() error {
		err := proto.Unmarshal(bs, msg)
		return errors.Wrap(err, "unmarshal protobuf")
	})
}

func (u *unmarshaller) do(op func() error) {
	if u.Err != nil {
		return
	}
	u.Err = op()
}
