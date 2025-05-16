package addmanual

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carlosrabelo/tatuscan/tools/tools/internal/apiclient"
)

func TestGenerateMachineID(t *testing.T) {
	a := GenerateMachineID("host", "")
	b := GenerateMachineID("host", "salt")
	if a == b || len(a) != 64 {
		t.Fatalf("ids: %s %s", a, b)
	}
	if GenerateMachineID("host", "") != a {
		t.Fatal("expected stable hash")
	}
}

func TestNormalizeActivation(t *testing.T) {
	if got := NormalizeActivation("15/01/2024"); got != "2024-01-15" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeActivation("raw"); got != "raw" {
		t.Fatalf("passthrough: %q", got)
	}
}

func TestPayloadAndSend(t *testing.T) {
	p := Payload(Entry{Hostname: "h1", OS: "Chrome OS", IP: "0.0.0.0", Activation: "2024-01-02"})
	if p["machine_id"] == "" || p["computer_activation"] != "2024-01-02" {
		t.Fatalf("payload: %v", p)
	}

	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "ok"})
	}))
	t.Cleanup(srv.Close)

	err := Send(context.Background(), apiclient.New(srv.URL, "tok"), Entry{
		Hostname: "h1", OS: "Chrome OS", IP: "1.2.3.4", MachineID: "fixed",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if got["machine_id"] != "fixed" || got["hostname"] != "h1" {
		t.Fatalf("posted: %v", got)
	}
	if err := Send(context.Background(), apiclient.New(srv.URL, "tok"), Entry{Hostname: "x"}, true); err != nil {
		t.Fatal(err)
	}
}
