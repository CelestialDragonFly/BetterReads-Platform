package middleware

import (
	"context"
	"net/http"
	"strings"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/auth"
	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

// Authentication is a middleware function that validates the Authorization header in incoming requests.
// It extracts the bearer token, verifies it using the provided Authenticator, and sets the authenticated
// user ID in the request headers if successful. If authentication fails, it responds with an error.
//
// Parameters:
//   - auth: An implementation of the Authenticator interface used to verify the ID token.
//
// Returns:
//   - A betterreads.StrictMiddlewareFunc that wraps an HTTP handler function, enforcing authentication.
func Authentication(authn auth.Authenticator) betterreads.StrictMiddlewareFunc {
	return func(f strictnethttp.StrictHTTPHandlerFunc, _ string) strictnethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			token, verifyError := authn.VerifyIDToken(ctx, strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1))
			if verifyError != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusPreconditionFailed)
				_, _ = w.Write([]byte(`{"code": "TBD", "message": "invalid authorization"}`))
				//nolint: nilerr // error in this context is reserved for internal server errors and our open API doesn't define middleware yet.
				return nil, nil
			}

			r.Header.Set("userid", token.UserID)
			return f(ctx, w, r, request)
		}
	}
}
