//go:build darwin

package internal

// ensureSingleInstance is a no-op on macOS
func EnsureSingleInstance() {
	Log.Debug("Single instance check is no-op on macOS; launchd manages the unique instance")
	// No-op: launchd manages the unique instance on macOS
}
