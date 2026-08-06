package telemetry

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/balramadan/gopherqueue"
)

type SSEBroadcaster struct {
	store     gopherqueue.Storage
	queueName string
}

func NewSSEBroadcaster(store gopherqueue.Storage, queueName string) *SSEBroadcaster {
	if queueName == "" {
		queueName = "default"
	}
	return &SSEBroadcaster{
		store:     store,
		queueName: queueName,
	}
}

func (b *SSEBroadcaster) StreamMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "gopherqueue/telemetry: streaming unsupported", http.StatusInternalServerError)
		return
	}

	clientAddr := r.RemoteAddr
	log.Printf("[SSE Telemetry] Client connected: %s", clientAddr)

	ctx := r.Context()

	// Initial instant flush upon connection
	if metrics, err := CollectMetrics(ctx, b.store, b.queueName); err == nil {
		if dataBytes, err := json.Marshal(metrics); err == nil {
			fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
			flusher.Flush()
		}
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[SSE Telemetry] Client disconnected: %s", clientAddr)
			return
		case <-ticker.C:
			metrics, err := CollectMetrics(ctx, b.store, b.queueName)
			if err != nil {
				log.Printf("[SSE Telemetry] Error collecting metrics: %v", err)
				continue
			}

			dataBytes, err := json.Marshal(metrics)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
			flusher.Flush()
		}
	}
}

func (b *SSEBroadcaster) HTTPHandler() http.HandlerFunc {
	return b.StreamMetrics
}
