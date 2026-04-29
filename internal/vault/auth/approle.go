package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/AndreyEremin02/vault-env/internal/config"
	vault "github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

func init() {
	Register("approle", func(cfg *config.Config) (Authenticator, error) {
		if cfg.RoleID == "" || cfg.SecretID == "" {
			return nil, errors.New("--role-id (or VAULT_ROLE_ID) and --secret-id (or VAULT_SECRET_ID) are required for approle auth")
		}
		return &appRoleAuth{roleID: cfg.RoleID, secretID: cfg.SecretID}, nil
	}, "--role-id", "--secret-id")
}

type appRoleAuth struct {
	roleID   string
	secretID string
}

func (a *appRoleAuth) Authenticate(ctx context.Context, client *vault.Client) error {
	resp, err := client.Auth.AppRoleLogin(ctx, schema.AppRoleLoginRequest{
		RoleId:   a.roleID,
		SecretId: a.secretID,
	})
	if err != nil {
		return fmt.Errorf("approle login: %w", err)
	}
	if resp.Auth == nil {
		return fmt.Errorf("approle login returned no auth info")
	}
	return client.SetToken(resp.Auth.ClientToken)
}
