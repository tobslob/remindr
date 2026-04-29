package utils

import (
	"context"

	"github.com/tobslob/remindr/internal/requestctx"
	"github.com/tobslob/remindr/internal/store"
)

func GetUserFromContext(ctx context.Context) *store.User {
	user, _ := requestctx.User(ctx)
	return user
}
