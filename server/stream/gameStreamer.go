package stream

import (
	"DurakGo/server"
	"fmt"
	"net/http"
)

type GameStreamer struct {
	streamer   SSEStreamer
	playerName string
}

func NewGameStreamer() (gameStreamer *GameStreamer) {
	gameStreamer = &GameStreamer{
		streamer: *NewSSEStreamer(),
	}

	return gameStreamer
}

func (this *GameStreamer) RegisterClient(w *http.ResponseWriter, r *http.Request) chan server.JSONResponseData {
	return RegisterClient(w, r)
}

func (this *GameStreamer) StreamLoop(w *http.ResponseWriter, messageChan chan server.JSONResponseData,
	customizeDataFunc func(server.JSONResponseData) server.JSONResponseData) {
	flusher, ok := (*w).(http.Flusher)

	if !ok {
		http.Error(*w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Make sure to close connection
	defer removeClient(messageChan)

	for {
		// Write to the ResponseWriter
		// Server Sent Events compatible
		originalData := <- messageChan
		customizedData := customizeDataFunc(originalData)

		if _, err := fmt.Fprintf(*w, "%s", convertToString(customizedData)); err != nil {
			http.Error(*w, "Problem writing data to event", http.StatusInternalServerError)
			return
		}

		// Flush the data immediately instead of buffering it for later.
		flusher.Flush()
	}
}

func (this *GameStreamer) Publish(respData server.JSONResponseData) {

	Publish(respData)

}
