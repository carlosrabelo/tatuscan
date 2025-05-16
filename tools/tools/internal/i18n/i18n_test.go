package i18n

import "testing"

func TestFromEnv(t *testing.T) {
	t.Setenv("TATUSCAN_LANG", "")
	if FromEnv().T("add.hostname") != "--hostname is required" {
		t.Fatal(FromEnv().T("add.hostname"))
	}
	t.Setenv("TATUSCAN_LANG", "pt")
	if FromEnv().T("add.hostname") != "--hostname é obrigatório" {
		t.Fatal(FromEnv().T("add.hostname"))
	}
}
