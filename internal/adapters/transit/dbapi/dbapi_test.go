package dbapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDBApiProvider(t *testing.T) {
	mockResponse := `<?xml version="1.0" encoding="UTF-8"?>
<timetable station="München Hbf (tief)">
    <s id="-123456789">
        <ar pt="2603211800" ct="2603211805" l="S1"/>
    </s>
</timetable>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	provider := NewProvider(server.URL, "test-client", "test-secret", []string{"8098263"})
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	records, logs, err := provider.FetchDelays(ctx)
	if err != nil {
		t.Fatalf("FetchDelays failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
}
