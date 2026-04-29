package secret

import "context"

// Backend reads secrets from a source and returns them as key-value pairs.
type Backend interface {
	Load(ctx context.Context, path string) (map[string]string, error)
}
