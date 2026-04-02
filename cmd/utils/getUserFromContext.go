package utils

import (
	"context"

	"github.com/tobslob/todoApp/internal/requestctx"
	"github.com/tobslob/todoApp/internal/store"
)

func GetUserFromContext(ctx context.Context) *store.User {
	user, _ := requestctx.User(ctx)
	return user
}
