//go:build !windows

package credentials

// Add Keychain/Secret Service factories with OS-specific build tags when their
// backends are implemented. Unsupported systems still allow session-only use.
func New(string) Vault { return unsupportedVault{} }
