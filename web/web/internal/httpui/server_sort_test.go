package httpui

import (
	"testing"
	"time"

	"github.com/carlosrabelo/tatuscan/web/web/internal/apiclient"
)

func TestSortReportRowsStatus(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	old := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339Nano)
	rows := []reportRow{
		{Machine: apiclient.Machine{Hostname: "offline-a", UpdatedAt: &old}, Online: false},
		{Machine: apiclient.Machine{Hostname: "online-z", UpdatedAt: &now}, Online: true},
		{Machine: apiclient.Machine{Hostname: "offline-b", UpdatedAt: &old}, Online: false},
	}
	sortReportRows(rows, "status", "asc")
	if rows[0].Online || !rows[2].Online {
		t.Fatalf("asc should put offline first: %+v %+v %+v", rows[0], rows[1], rows[2])
	}
	sortReportRows(rows, "status", "desc")
	if !rows[0].Online || rows[2].Online {
		t.Fatalf("desc should put online first: %+v %+v %+v", rows[0], rows[1], rows[2])
	}
}
