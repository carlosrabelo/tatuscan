package i18n

import "testing"

func TestParseAndT(t *testing.T) {
	if Parse("pt-BR") != PT {
		t.Fatal("pt-BR")
	}
	if Parse("en-US") != EN {
		t.Fatal("en-US")
	}
	if New(EN).T("error.api_down") == "error.api_down" {
		t.Fatal("missing error.api_down")
	}
}
