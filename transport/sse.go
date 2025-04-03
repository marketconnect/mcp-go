package transport

import (
	"encoding/json"
	"net/http"
)

type SSETransport struct {
	writer   http.ResponseWriter
	flusher  http.Flusher
	sendChan chan []byte
}

// NewSSETransport initializes a new SSETransport instance for Server-Sent Events (SSE).
// It sets the necessary headers on the http.ResponseWriter to establish an SSE stream,
// and starts a background goroutine to handle sending event data to the client.
func NewSSETransport(w http.ResponseWriter) *SSETransport {
	// Set headers to establish SSE stream
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, _ := w.(http.Flusher)
	transport := &SSETransport{
		writer:   w,
		flusher:  flusher,
		sendChan: make(chan []byte, 10),
	}
	// Force sending Headers to the client immediately
	transport.flusher.Flush()
	// Start a goroutine to receive data from the channel and send it to the client
	go func() {
		for data := range transport.sendChan {
			// Write event data in SSE format
			_, _ = transport.writer.Write([]byte("data: "))
			_, _ = transport.writer.Write(data)
			_, _ = transport.writer.Write([]byte("\n\n"))
			// Flush the data to the client
			transport.flusher.Flush()
		}
	}()
	return transport
}

// Send marshals the given message to JSON and sends it to the client over the SSE stream.
// If marshaling fails, an error is returned.

func (sse *SSETransport) Send(message interface{}) error {
	//  Marshal the message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	//
	sse.sendChan <- jsonData
	return nil
}

// Close closes the underlying channel used to send events to the client, and
// will cause the background goroutine started by NewSSETransport to exit.
func (sse *SSETransport) Close() {
	close(sse.sendChan)
}
