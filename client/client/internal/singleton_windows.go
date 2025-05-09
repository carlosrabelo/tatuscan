//go:build windows

package internal

// ensureSingleInstance checks if another instance is running
func EnsureSingleInstance() {
	Log.Debug("Single instance check temporarily disabled on Windows")
	// No-op: Temporarily disabled for debugging on Windows 7
}
