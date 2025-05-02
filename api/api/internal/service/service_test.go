package service_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlosrabelo/tatuscan/api/api/internal/service"
	"github.com/carlosrabelo/tatuscan/api/api/internal/store"
)

func newTestService(t *testing.T) *service.Service {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc, err := service.New(st, "America/Cuiaba")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return svc
}

func sampleData(id, host string) map[string]any {
	return map[string]any{
		"machine_id":      id,
		"hostname":        host,
		"ip":              "192.168.1.10",
		"os":              "linux",
		"os_version":      "Ubuntu 22.04",
		"cpu_percent":     12.5,
		"memory_total_mb": 8192,
		"memory_used_mb":  2048,
	}
}

func TestCreateAndList(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, created, err := svc.CreateOrUpdate(ctx, sampleData("m1", "zebra"))
	if err != nil || !created {
		t.Fatalf("create: created=%v err=%v", created, err)
	}
	_, created, err = svc.CreateOrUpdate(ctx, sampleData("m2", "alpha"))
	if err != nil || !created {
		t.Fatalf("create2: created=%v err=%v", created, err)
	}

	list, err := svc.ListAll(ctx, "hostname", "asc")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].Hostname != "alpha" || list[1].Hostname != "zebra" {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestUpsertUpdates(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	_, _, err := svc.CreateOrUpdate(ctx, sampleData("m1", "host1"))
	if err != nil {
		t.Fatal(err)
	}
	data := sampleData("m1", "host1-updated")
	data["cpu_percent"] = 50.0
	inv, created, err := svc.CreateOrUpdate(ctx, data)
	if err != nil || created {
		t.Fatalf("upsert: created=%v err=%v", created, err)
	}
	if inv.Hostname != "host1-updated" || inv.CPUPercent != 50.0 {
		t.Fatalf("unexpected inv: %+v", inv)
	}
	if inv.UpdatedAt == nil {
		t.Fatal("expected updated_at")
	}
}

func TestCreateTruncatesUTF8Hostname(t *testing.T) {
	svc := newTestService(t)
	data := sampleData("m1", strings.Repeat("á", 120))
	inv, _, err := svc.CreateOrUpdate(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	if got := []rune(inv.Hostname); len(got) != 100 {
		t.Fatalf("hostname runes=%d want 100", len(got))
	}
}

func TestMissingRequired(t *testing.T) {
	svc := newTestService(t)
	_, _, err := svc.CreateOrUpdate(context.Background(), map[string]any{"hostname": "x"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	se, ok := err.(*service.Error)
	if !ok || se.StatusCode != 400 {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRejectNullRequired(t *testing.T) {
	svc := newTestService(t)
	data := sampleData("m1", "host1")
	data["hostname"] = nil
	if _, _, err := svc.CreateOrUpdate(context.Background(), data); err == nil {
		t.Fatal("expected validation error for null hostname")
	}
	data = sampleData("m1", "   ")
	if _, _, err := svc.CreateOrUpdate(context.Background(), data); err == nil {
		t.Fatal("expected validation error for empty hostname")
	}
}

func TestUpsertPreservesOptionalFields(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	created, _, err := svc.CreateOrUpdate(ctx, sampleData("m1", "host1"))
	if err != nil {
		t.Fatal(err)
	}
	if created.OSVersion == nil || *created.OSVersion != "Ubuntu 22.04" {
		t.Fatalf("os_version: %v", created.OSVersion)
	}
	if created.MemoryUsedMB == nil || *created.MemoryUsedMB != 2048 {
		t.Fatalf("memory_used_mb: %v", created.MemoryUsedMB)
	}

	partial := map[string]any{
		"machine_id": "m1", "hostname": "host1", "ip": "192.168.1.10",
		"os": "linux", "cpu_percent": 20.0, "memory_total_mb": 8192,
	}
	inv, createdFlag, err := svc.CreateOrUpdate(ctx, partial)
	if err != nil || createdFlag {
		t.Fatalf("upsert: created=%v err=%v", createdFlag, err)
	}
	if inv.OSVersion == nil || *inv.OSVersion != "Ubuntu 22.04" {
		t.Fatalf("os_version cleared: %v", inv.OSVersion)
	}
	if inv.MemoryUsedMB == nil || *inv.MemoryUsedMB != 2048 {
		t.Fatalf("memory_used_mb cleared: %v", inv.MemoryUsedMB)
	}
}
