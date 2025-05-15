package apiclient

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("TATUSCAN_URL=http://from-file:8040\nEXISTING=file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EXISTING", "env")
	if err := LoadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("TATUSCAN_URL"); got != "http://from-file:8040" {
		t.Fatalf("url: %s", got)
	}
	if got := os.Getenv("EXISTING"); got != "env" {
		t.Fatalf("must not overwrite: %s", got)
	}
	if err := LoadDotEnv(filepath.Join(dir, "missing.env")); err != nil {
		t.Fatal(err)
	}
}
