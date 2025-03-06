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

func NewFirebaseAuth(cfg Config) (*Firebase, error) {
	opt := option.WithCredentialsFile(cfg.FirebaseServiceAccount)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %w", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}
	return &Firebase{
		AuthClient: authClient,
	}, nil
}

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
