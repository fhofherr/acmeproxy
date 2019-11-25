package auth

import (
	"context"

	"github.com/fhofherr/acmeproxy/pkg/errors"
)

type ctxkey int

const (
	keyClaims ctxkey = iota
)

// AddClaimsToContext adds the passed claims to the context ctx.
func AddClaimsToContext(ctx context.Context, cs *Claims) context.Context {
	return context.WithValue(ctx, keyClaims, cs)
}

// ClaimsFromContext returns claims stored in the context. The second return
// value is false if no claims were found in the context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	cs, ok := ctx.Value(keyClaims).(*Claims)
	return cs, ok
}

// CheckRoles checks if the passed context contains claims  and if those claims
// contain at least one of the expected roles.
func CheckRoles(ctx context.Context, roles ...Role) error {
	const op errors.Op = "auth/CheckRoles"

	cs, ok := ClaimsFromContext(ctx)
	if !ok {
		return errors.New(op, errors.Unspecified, "no claims in context")
	}
	return errors.Wrap(cs.CheckRoles(roles...), op)
}
