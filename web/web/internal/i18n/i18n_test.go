package i18n

import "testing"

func TestParseAndT(t *testing.T) {
	if Parse("pt-BR") != PT {
		t.Fatal("pt-BR")
	}
	if Parse("en-US") != EN {
		t.Fatal("en-US")
	}
	if New(PT).T("nav.report") != "Relatório" {
		t.Fatal(New(PT).T("nav.report"))
	}
	if New(EN).T("nav.report") != "Report" {
		t.Fatal(New(EN).T("nav.report"))
	}
	if New(PT).T("home.threshold", "2h0m0s") != "2h0m0s sem check-in" {
		t.Fatal(New(PT).T("home.threshold", "2h0m0s"))
	}
}
