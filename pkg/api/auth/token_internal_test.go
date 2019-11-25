package auth

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestAlgorithm_MethodOk(t *testing.T) {
	tests := []struct {
		alg    Algorithm
		method jwt.SigningMethod
		ok     bool
	}{
		{
			alg:    Algorithm(-1),
			method: jwt.SigningMethodES256,
		},
		{
			alg:    Algorithm(-1),
			method: jwt.SigningMethodES384,
		},
		{
			alg:    Algorithm(-1),
			method: jwt.SigningMethodES512,
		},
		{
			alg:    ES256,
			method: jwt.SigningMethodES256,
			ok:     true,
		},
		{
			alg:    ES256,
			method: jwt.SigningMethodES384,
		},
		{
			alg:    ES256,
			method: jwt.SigningMethodES512,
		},
		{
			alg:    ES512,
			method: jwt.SigningMethodES256,
		},
		{
			alg:    ES512,
			method: jwt.SigningMethodES384,
		},
		{
			alg:    ES512,
			method: jwt.SigningMethodES512,
			ok:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			ok := tt.alg.methodOk(tt.method)
			assert.Equal(t, tt.ok, ok)
		})
	}
}
