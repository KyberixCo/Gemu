//go:build windows

package credentials

// New selects the native backend for this build and scopes it to a Gemu profile.
func New(dir string) Vault { return newFileStore(dir) }
