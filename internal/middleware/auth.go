package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/auth"
	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

type Authenticator interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

func Authentication(auth Authenticator) betterreads.StrictMiddlewareFunc {
	return func(f strictnethttp.StrictHTTPHandlerFunc, _ string) strictnethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusPreconditionFailed)
				return nil, fmt.Errorf("missing authorization")
			}

			token, verifyError := auth.VerifyIDToken(ctx, strings.Replace(authHeader, "Bearer ", "", 1))
			if verifyError != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return nil, fmt.Errorf("unauthorized")
			}

			r.Header.Set("userid", token.UserID)
			return f(ctx, w, r, request)
		}
	}
}
