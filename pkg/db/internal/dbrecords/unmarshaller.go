package dbrecords

import (
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
)

// UnmarshalBinary unmarshals the passed byte slice into v which should be a pointer
// to a passed domain object.
func UnmarshalBinary(bs []byte, v interface{}) error {
	u := &BinaryUnmarshaller{V: v}
	return u.UnmarshalBinary(bs)
}

// BinaryUnmarshaller wraps a target value V and reads its contents from a binary
// representation.
//
// Once the Umarshaller has been used it must not be used again.
type BinaryUnmarshaller struct {
	V   interface{}
	err error
}

// UnmarshalBinary creates a domain object from the passed bytes.
func (u *BinaryUnmarshaller) UnmarshalBinary(bs []byte) error {
	const op errors.Op = "dbrecords/binaryUnmarshaller.UnmarshalBinary"

	switch obj := u.V.(type) {
	case *acme.User:
		u.unmarshalACMEUser(bs, obj)
	case *acme.Domain:
		u.unmarshalACMEDomain(bs, obj)
	case *uuid.UUID:
		u.unmarshalUUID(bs, obj)
	case *string:
		*obj = string(bs)
	default:
		return errors.New(op, fmt.Sprintf("unsupported type: %T", u.V))
	}
	if u.err != nil {
		return errors.New(op, fmt.Sprintf("unmarshal type: %T", u.V), u.err)
	}
	return nil
}

func (u *BinaryUnmarshaller) unmarshalUUID(bs []byte, id *uuid.UUID) {
	const op errors.Op = "dbrecords/binaryUnmarshaller.unmarshalUUID"

	u.do(func() error {
		err := id.UnmarshalBinary(bs)
		if err != nil {
			return errors.New(op, "unmarshal uuid", err)
		}
		return nil
	})
}

func (u *BinaryUnmarshaller) unmarshalACMEUser(bs []byte, user *acme.User) {
	const op errors.Op = "dbrecords/binaryUnmarshaller.unmarshalACMEUser"

	u.do(func() error {
		if user == nil {
			return errors.New(op, "user must not be nil")
		}
		var (
			rec User
			id  uuid.UUID
		)
		u.unmarshalPB(bs, &rec)
		u.unmarshalUUID(rec.Id, &id)
		key := u.unmarshalPrivateKey(keyType(rec.AccountKey.KeyType), rec.AccountKey.KeyBytes)
		user.AccountURL = rec.AccountURL
		user.ID = id
		user.Key = key
		return nil
	})
}

func (u *BinaryUnmarshaller) unmarshalPrivateKey(kt keyType, bs []byte) crypto.PrivateKey {
	const op errors.Op = "dbrecords/binaryUnmarshaller.unmarshalPrivateKey"

	var pk crypto.PrivateKey
	u.do(func() error {
		var err error

		switch kt {
		case ecdsa:
			pk, err = x509.ParseECPrivateKey(bs)
			if err != nil {
				err = errors.New(op, "unmarshal ECDSA private key", err)
			}
		case rsa:
			pk, err = x509.ParsePKCS1PrivateKey(bs)
			if err != nil {
				err = errors.New(op, "parse RSA private key", err)
			}
		default:
			err = errors.New(op, fmt.Sprintf("unknown key type: %v", kt))
		}
		return err
	})
	return pk
}

func (u *BinaryUnmarshaller) unmarshalACMEDomain(bs []byte, domain *acme.Domain) {
	const op errors.Op = "dbrecords/binaryUnmarshaller.unmarshalACMEDomain"

	u.do(func() error {
		if domain == nil {
			return errors.New(op, "domain must not be nil")
		}
		var (
			rec    Domain
			userID uuid.UUID
		)
		u.unmarshalPB(bs, &rec)
		u.unmarshalUUID(rec.UserID, &userID)
		domain.UserID = userID
		domain.Name = rec.Name
		domain.Certificate = rec.CertificatePEM
		domain.PrivateKey = rec.PrivateKeyPEM
		return nil
	})
}

func (u *BinaryUnmarshaller) unmarshalPB(bs []byte, msg proto.Message) {
	const op errors.Op = "dbrecords/binaryUnmarshaller.unmarshalPB"

	u.do(func() error {
		err := proto.Unmarshal(bs, msg)
		if err != nil {
			return errors.New(op, "unmarshal protobuf", err)
		}
		return nil
	})
}

func (u *BinaryUnmarshaller) do(op func() error) {
	if u.err != nil {
		return
	}
	u.err = op()
}
