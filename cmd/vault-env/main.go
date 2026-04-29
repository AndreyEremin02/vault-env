package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AndreyEremin02/vault-env/internal/config"
	"github.com/AndreyEremin02/vault-env/internal/executor"
	"github.com/AndreyEremin02/vault-env/internal/logger"
	vaultpkg "github.com/AndreyEremin02/vault-env/internal/vault"
	"github.com/AndreyEremin02/vault-env/internal/vault/auth"
	"github.com/AndreyEremin02/vault-env/internal/vault/secret"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

var VersionString string

const (
	exitAppError     = 1
	exitAuthError    = 2
	exitSecretsError = 3
	exitCmdError     = 4
)

func main() {
	app := &cli.Command{
		Name:      "vault-env",
		Usage:     "Inject HashiCorp Vault secrets as environment variables and run a command",
		UsageText: "vault-env [options] command [arguments]",
		Version:   VersionString,
		Flags:     cliFlags(),
		Action:    action,
		ExitErrHandler: func(_ context.Context, _ *cli.Command, _ error) {},
		OnUsageError: func(ctx context.Context, cmd *cli.Command, err error, _ bool) error {
			_ = cli.ShowAppHelp(cmd)
			return err
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		os.Exit(exitAppError)
	}
}

func cliFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"a"},
			Usage:   "Vault server address",
			Sources: cli.EnvVars("VAULT_ADDR"),
		},
		&cli.StringFlag{
			Name:    "auth-method",
			Aliases: []string{"m"},
			Usage:   "Authentication method: " + auth.SupportedMethods(),
			Sources: cli.EnvVars("VAULT_AUTH_METHOD"),
		},
		&cli.StringFlag{
			Name:    "token",
			Aliases: []string{"t"},
			Usage:   "Vault token (token auth)",
			Sources: cli.EnvVars("VAULT_TOKEN"),
		},
		&cli.StringFlag{
			Name:    "role-id",
			Usage:   "AppRole role ID",
			Sources: cli.EnvVars("VAULT_ROLE_ID"),
		},
		&cli.StringFlag{
			Name:    "secret-id",
			Usage:   "AppRole secret ID",
			Sources: cli.EnvVars("VAULT_SECRET_ID"),
		},
		&cli.StringSliceFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Usage:   "KV v2 secret path (repeatable)",
			Sources: cli.EnvVars("VAULT_PATHS"),
		},
		&cli.StringFlag{
			Name:    "mount",
			Usage:   "KV v2 mount path",
			Sources: cli.EnvVars("VAULT_MOUNT"),
			Value:   "kv",
		},
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "Enable debug logging",
			Sources: cli.EnvVars("VAULT_ENV_DEBUG"),
		},
		&cli.BoolFlag{
			Name:    "silent",
			Usage:   "Silence all log output",
			Sources: cli.EnvVars("VAULT_ENV_SILENT"),
		},
		&cli.BoolFlag{
			Name:    "no-expand",
			Usage:   "Do not expand env variable references in secret values",
			Sources: cli.EnvVars("VAULT_ENV_NO_EXPAND"),
		},
		&cli.BoolFlag{
			Name:    "tls-skip-verify",
			Usage:   "Skip TLS certificate verification",
			Sources: cli.EnvVars("VAULT_SKIP_VERIFY"),
		},
	}
}

func action(ctx context.Context, cmd *cli.Command) error {
	logger.Setup(cmd.Bool("debug"), cmd.Bool("silent"))

	if cmd.NArg() == 0 {
		_ = cli.ShowAppHelp(cmd)
		return cli.Exit("", exitCmdError)
	}

	cfg := &config.Config{
		VaultAddr:     cmd.String("address"),
		AuthMethod:    cmd.String("auth-method"),
		Token:         cmd.String("token"),
		RoleID:        cmd.String("role-id"),
		SecretID:      cmd.String("secret-id"),
		SecretPaths:   cmd.StringSlice("path"),
		Mount:         cmd.String("mount"),
		Debug:         cmd.Bool("debug"),
		Silent:        cmd.Bool("silent"),
		NoExpand:      cmd.Bool("no-expand"),
		TLSSkipVerify: cmd.Bool("tls-skip-verify"),
	}

	if err := validateConfig(cfg); err != nil {
		log.Error(err.Error())
		return cli.Exit("", exitAppError)
	}

	authenticator, err := auth.New(cfg.AuthMethod, cfg)
	if err != nil {
		log.Error(err.Error())
		return cli.Exit("", exitAuthError)
	}

	client, err := vaultpkg.NewClient(ctx, cfg, authenticator)
	if err != nil {
		log.WithError(err).Error("vault client initialization failed")
		return cli.Exit("", exitAuthError)
	}

	loader := secret.NewKV2Loader(client.Raw(), cfg.Mount)

	if err := injectSecrets(ctx, loader, cfg); err != nil {
		log.WithError(err).Error("failed to load secrets")
		return cli.Exit("", exitSecretsError)
	}

	return executor.Execute(cmd.Args().First(), cmd.Args().Slice()[1:])
}

func validateConfig(cfg *config.Config) error {
	if cfg.VaultAddr == "" {
		return errors.New("--address (or VAULT_ADDR) is required")
	}
	if cfg.AuthMethod == "" {
		return fmt.Errorf("--auth-method (or VAULT_AUTH_METHOD) is required (supported: %s)", auth.SupportedMethods())
	}
	if len(cfg.SecretPaths) == 0 {
		return errors.New("at least one --path is required")
	}
	return nil
}

func injectSecrets(ctx context.Context, loader secret.Backend, cfg *config.Config) error {
	for _, path := range cfg.SecretPaths {
		secrets, err := loader.Load(ctx, path)
		if err != nil {
			return err
		}
		for k, v := range secrets {
			log.WithField("key", k).Debug("injecting env var")
			if err := os.Setenv(k, v); err != nil {
				return fmt.Errorf("set env var %s: %w", k, err)
			}
		}
	}

	if !cfg.NoExpand {
		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			if err := os.Setenv(pair[0], os.Expand(pair[1], os.Getenv)); err != nil {
				return fmt.Errorf("expand env var %s: %w", pair[0], err)
			}
		}
	}

	return nil
}
