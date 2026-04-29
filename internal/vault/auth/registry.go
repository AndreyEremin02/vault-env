package auth

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AndreyEremin02/vault-env/internal/config"
	log "github.com/sirupsen/logrus"
)

type Factory func(cfg *config.Config) (Authenticator, error)

type registration struct {
	factory Factory
	claims  []string
}

var registry = map[string]registration{}

// authFields maps flag name → getter for auth-specific config fields.
// Add an entry here when a new auth-specific flag is introduced.
var authFields = map[string]func(*config.Config) bool{
	"--token":     func(cfg *config.Config) bool { return cfg.Token != "" },
	"--role-id":   func(cfg *config.Config) bool { return cfg.RoleID != "" },
	"--secret-id": func(cfg *config.Config) bool { return cfg.SecretID != "" },
}

// Register registers an auth method with the fields it consumes.
// Any auth-specific field not listed in claims will trigger a warning at runtime.
func Register(name string, f Factory, claims ...string) {
	registry[name] = registration{factory: f, claims: claims}
}

func New(method string, cfg *config.Config) (Authenticator, error) {
	reg, ok := registry[method]
	if !ok {
		return nil, fmt.Errorf("unknown auth method %q (supported: %s)", method, SupportedMethods())
	}
	warnUnused(method, cfg, reg.claims)
	return reg.factory(cfg)
}

func SupportedMethods() string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func warnUnused(method string, cfg *config.Config, claims []string) {
	claimed := make(map[string]bool, len(claims))
	for _, c := range claims {
		claimed[c] = true
	}
	for flag, isSet := range authFields {
		if !claimed[flag] && isSet(cfg) {
			log.Warnf("%s is ignored with %s auth", flag, method)
		}
	}
}
