package auth

import (
	"context"
	"errors"

	"github.com/AndreyEremin02/vault-env/internal/config"
	vault "github.com/hashicorp/vault-client-go"
)

func init() {
	Register("token", func(cfg *config.Config) (Authenticator, error) {
		if cfg.Token == "" {
			return nil, errors.New("--token (or VAULT_TOKEN) is required for token auth")
		}
		return &tokenAuth{token: cfg.Token}, nil
	}, "--token")
}

type tokenAuth struct {
	token string
}

func (t *tokenAuth) Authenticate(_ context.Context, client *vault.Client) error {
	return client.SetToken(t.token)
}
