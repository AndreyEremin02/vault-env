package auth

import (
	"context"

	vault "github.com/hashicorp/vault-client-go"
)

// Authenticator authenticates against Vault and sets the token on the client.
type Authenticator interface {
	Authenticate(ctx context.Context, client *vault.Client) error
}
