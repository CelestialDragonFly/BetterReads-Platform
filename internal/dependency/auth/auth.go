package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

type AuthUser struct {
	AuthTime int64  `json:"auth_time"`
	Issuer   string `json:"iss"`
	Audience string `json:"aud"`
	Expires  int64  `json:"exp"`
	IssuedAt int64  `json:"iat"`
	Subject  string `json:"sub,omitempty"`
	UserID   string `json:"uid,omitempty"`
}

func NewAuth(filePath string) (*auth.Client, error) {
	opt := option.WithCredentialsFile(filePath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase authentication: %w", err)
	}

	auth, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase authentication: %w", err)
	}
	return auth, nil
}
