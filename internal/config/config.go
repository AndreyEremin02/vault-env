package config

type Config struct {
	VaultAddr     string
	AuthMethod    string
	Token         string
	RoleID        string
	SecretID      string
	SecretPaths   []string
	Mount         string
	Debug         bool
	Silent        bool
	NoExpand      bool
	TLSSkipVerify bool
}
