package requestctx

import (
	"context"

	"github.com/tobslob/todoApp/internal/store"
)

type userKey struct{}

var currentUserKey userKey

func WithUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, currentUserKey, user)
}

func User(ctx context.Context) (*store.User, bool) {
	user, ok := ctx.Value(currentUserKey).(*store.User)
	return user, ok
}
