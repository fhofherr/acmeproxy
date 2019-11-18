package auth

import "context"

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
