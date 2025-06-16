package headers

import (
	"context"
)

type ContextKey string

const UserIDContextKey ContextKey = "user_id"

func GetUserID(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(UserIDContextKey).(string)
	return u, ok
}
