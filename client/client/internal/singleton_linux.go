//go:build linux

package internal

// ensureSingleInstance is a no-op on Linux
func EnsureSingleInstance() {
	Log.Debug("Single instance check is no-op on Linux; systemd manages the unique instance")
	// No-op: systemd manages the unique instance on Linux
}
