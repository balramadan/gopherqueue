package telemetry_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/balramadan/gopherqueue/telemetry"
)

func TestSSEBroadcaster_StreamMetrics(t *testing.T) {
	store := &mockTelemetryStorage{
		pending:    3,
		processing: 1,
	}

	broadcaster := telemetry.NewSSEBroadcaster(store, "default")
	ts := httptest.NewServer(broadcaster.HTTPHandler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", ts.URL, nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	broadcaster.StreamMetrics(w, req)

	res := w.Result()
	if res.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", res.Header.Get("Content-Type"))
	}

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "data:") || !strings.Contains(bodyStr, `"pending":3`) {
		t.Errorf("expected SSE data format with pending:3, got body:\n%s", bodyStr)
	}
}
