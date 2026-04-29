package secret

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault-client-go"
	log "github.com/sirupsen/logrus"
)

type KV2Loader struct {
	client *vault.Client
	mount  string
}

func NewKV2Loader(client *vault.Client, mount string) *KV2Loader {
	return &KV2Loader{client: client, mount: mount}
}

func (k *KV2Loader) Load(ctx context.Context, path string) (map[string]string, error) {
	log.WithField("path", path).Debug("reading kv2 secret")

	resp, err := k.client.Secrets.KvV2Read(ctx, path, vault.WithMountPath(k.mount))
	if err != nil {
		return nil, fmt.Errorf("read secret %q: %w", path, err)
	}

	result := make(map[string]string, len(resp.Data.Data))
	for key, val := range resp.Data.Data {
		switch v := val.(type) {
		case string:
			result[key] = v
		default:
			result[key] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}
