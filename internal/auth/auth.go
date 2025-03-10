package auth

import (
	"context"
)

type Token struct {
	AuthTime int64
	Issuer   string
	Audience string
	Expires  int64
	IssuedAt int64
	Subject  string
	UserID   string
	// FirebaseInfo represents the information about the sign-in event, including which auth provider
	// was used and provider-specific identity details.
	Info struct {
		SignInProvider string
		Tenant         string
		Identities     map[string]any
	}
	Claims map[string]any
}

// Authenticator defines an interface for verifying an ID token.
// Implementations of this interface should provide a method to validate
// the token and return the corresponding authentication token or an error.
type Authenticator interface {
	VerifyIDToken(ctx context.Context, idToken string) (*Token, error)
}
