package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/celestialdragonfly/betterreads-platform/internal/dependency/auth"
	iErrors "github.com/celestialdragonfly/betterreads-platform/internal/package/errors"
)

func (s *Server) AuthUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	authToken := r.Header.Get("Authentication")
	token, err := s.Firebase.VerifyIDToken(ctx, authToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		iErrors.NewHttpError(&w, "user is not aunthenticated", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("user-id", token.UID)
	json.NewEncoder(w).Encode(auth.AuthUser{
		AuthTime: token.AuthTime,
		Issuer:   token.Issuer,
		Audience: token.Audience,
		Expires:  token.Expires,
		IssuedAt: token.IssuedAt,
		Subject:  token.Subject,
		UserID:   token.UID,
	})
}
