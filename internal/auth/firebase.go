package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

type Config struct {
	FirebaseServiceAccount string
}

type Firebase struct {
	AuthClient *auth.Client
}

// Ensures Firebase implements the Authenticator interface.
var _ Authenticator = &Firebase{}

func NewFirebaseAuth(ctx context.Context, cfg Config) (*Firebase, error) {
	opt := option.WithCredentialsFile(cfg.FirebaseServiceAccount)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}
	return &Firebase{
		AuthClient: authClient,
	}, nil
}

func (fb *Firebase) VerifyIDToken(ctx context.Context, idToken string) (*Token, error) {
	token, err := fb.AuthClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("unable to verify firebase token. error: %w", err)
	}

	return &Token{
		AuthTime: token.AuthTime,
		Issuer:   token.Issuer,
		Audience: token.Audience,
		Expires:  token.Expires,
		IssuedAt: token.IssuedAt,
		Subject:  token.Subject,
		UserID:   token.UID,
		Info: struct {
			SignInProvider string
			Tenant         string
			Identities     map[string]any
		}{
			SignInProvider: token.Firebase.SignInProvider,
			Tenant:         token.Firebase.Tenant,
			Identities:     token.Firebase.Identities,
		},
		Claims: token.Claims,
	}, nil
}
