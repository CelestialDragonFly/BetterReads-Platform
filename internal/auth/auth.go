package auth

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
