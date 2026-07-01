package auth

import "context"

type contextKey struct{}

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, contextKey{}, claims)
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(contextKey{}).(*Claims)
	return claims, ok && claims != nil
}

func UserIDFromContext(ctx context.Context) string {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return ""
	}
	return claims.UserID
}
