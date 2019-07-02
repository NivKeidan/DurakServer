package stream

import (
	"DurakGo/server/httpPayloadObjects"
	"fmt"
	"net/http"
)

type SSEStreamer struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan httpPayloadObjects.JSONResponseData

	// New client connections
	newClients chan chan httpPayloadObjects.JSONResponseData

	// Closed client connections
	closingClients chan chan httpPayloadObjects.JSONResponseData

	// Client connections registry
	clients map[chan httpPayloadObjects.JSONResponseData]bool
}

func NewSSEStreamer() (streamer *SSEStreamer) {
	streamer = &SSEStreamer{
		Notifier:       make(chan httpPayloadObjects.JSONResponseData),
		newClients:     make(chan chan httpPayloadObjects.JSONResponseData),
		closingClients: make(chan chan httpPayloadObjects.JSONResponseData),
		clients:        make(map[chan httpPayloadObjects.JSONResponseData]bool),
	}

	go streamer.listen()

	return
}

func (this *SSEStreamer) RegisterClient(w *http.ResponseWriter, r *http.Request) chan httpPayloadObjects.JSONResponseData {

	this.addHeaders(w)

	// New client channels
	messageChan := make(chan httpPayloadObjects.JSONResponseData)
	this.newClients <- messageChan

	// Handle client-side disconnection
	ctx := r.Context()

	go func() {
		<-ctx.Done()
		this.removeClient(messageChan)
	}()

	return messageChan


}

func (this *SSEStreamer) StreamLoop(w *http.ResponseWriter, messageChan chan httpPayloadObjects.JSONResponseData) {

	flusher, ok := (*w).(http.Flusher)

	if !ok {
		http.Error(*w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Make sure to close connection
	defer this.removeClient(messageChan)

	for {
		// Write to the ResponseWriter
		// Server Sent Events compatible
		if _, err := fmt.Fprintf(*w, "%s", convertToString(<-messageChan)); err != nil {
			http.Error(*w, "Problem writing data to event", http.StatusInternalServerError)
			return
		}

		// Flush the data immediately instead of buffering it for later.
		flusher.Flush()
	}


}

func (this *SSEStreamer) listen() {
	for {
		select {
			case s := <-this.newClients:
				// A new client has connected.
				// Register their message channel
				this.clients[s] = true

			case s := <-this.closingClients:
				// A client has detached and we want to
				// stop sending them messages.
				delete(this.clients, s)

			case event := <-this.Notifier:
				// We got a new event from the outside!
				// Send event to all connected clients
				for clientMessageChan := range this.clients {
					clientMessageChan <- event
				}
		}
	}
}

func (this *SSEStreamer) Publish(respData httpPayloadObjects.JSONResponseData) {

	this.Notifier <- respData

}

func (this *SSEStreamer) addHeaders(writer *http.ResponseWriter) {
	(*writer).Header().Set("Content-Type", "text/event-stream")
	(*writer).Header().Set("Cache-Control", "no-cache")
	(*writer).Header().Set("Connection", "keep-alive")

}

func (this *SSEStreamer) removeClient(msgChan chan httpPayloadObjects.JSONResponseData) {
	this.closingClients <- msgChan
}