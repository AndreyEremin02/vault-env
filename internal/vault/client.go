package vault

import (
	"context"
	"fmt"

	"github.com/AndreyEremin02/vault-env/internal/config"
	"github.com/AndreyEremin02/vault-env/internal/vault/auth"
	vault "github.com/hashicorp/vault-client-go"
)

type Client struct {
	raw *vault.Client
}

func NewClient(ctx context.Context, cfg *config.Config, authenticator auth.Authenticator) (*Client, error) {
	opts := []vault.ClientOption{
		vault.WithAddress(cfg.VaultAddr),
	}
	if cfg.TLSSkipVerify {
		opts = append(opts, vault.WithTLS(vault.TLSConfiguration{
			InsecureSkipVerify: true,
		}))
	}

	c, err := vault.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("create vault client: %w", err)
	}

	if err := authenticator.Authenticate(ctx, c); err != nil {
		return nil, fmt.Errorf("vault authentication: %w", err)
	}

	return &Client{raw: c}, nil
}

func (c *Client) Raw() *vault.Client {
	return c.raw
}
