package stream

import (
	"DurakGo/output"
	"DurakGo/server/httpPayloadTypes"
	"fmt"
	"net/http"
	"time"
)

type GameStreamer struct {
	SSEStreamer
}

func NewGameStreamer(isAliveResp httpPayloadTypes.JSONResponseData, ttl int) (gameStreamer *GameStreamer) {
	output.Spit("Game streamer running")
	gameStreamer = &GameStreamer{
		*NewSSEStreamer(),
	}

	go func() {
		output.Spit("go routine - publish is alive game streamer - start")
		defer func() {
			output.Spit("go routine - publish is alive game streamer - ended")
		}()
		for {
			gameStreamer.Publish(isAliveResp)
			time.Sleep(time.Duration(ttl / 2) * time.Second)
		}
	}()

	return gameStreamer
}

func (this *GameStreamer) StreamLoop(w *http.ResponseWriter, messageChan chan httpPayloadTypes.JSONResponseData,
	r *http.Request, customizeDataFunc func(httpPayloadTypes.JSONResponseData) (httpPayloadTypes.JSONResponseData, error)) {

	flusher, ok := (*w).(http.Flusher)

	if !ok {
		http.Error(*w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Make sure to close connection
	defer this.removeClient(messageChan)

	// Handle client-side disconnection
	ctx := r.Context()

	for {
		select {
			case originalData := <-messageChan:
				customizedData, err := customizeDataFunc(originalData)
				if err != nil {
					http.Error(*w, "Problem writing data to event", http.StatusInternalServerError)
					return
				}

				if _, err := fmt.Fprintf(*w, "%s", convertToString(customizedData)); err != nil {
					http.Error(*w, "Problem writing data to event", http.StatusInternalServerError)
					return
				}

				// Flush the data immediately instead of buffering it for later.
				flusher.Flush()
			case <-ctx.Done():
				output.Spit("client closed connection to game streamer")
				return
		}
	}

}

func (this *GameStreamer) RemoveClient(msgChan chan httpPayloadTypes.JSONResponseData) {
	this.removeClient(msgChan)
}