package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SSEStreamer struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool
}

func NewStreamer() (streamer *SSEStreamer) {
	streamer = &SSEStreamer{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}

	go streamer.listen()

	return
}

func (this *SSEStreamer) registerClient(w *http.ResponseWriter, r *http.Request) chan []byte {

	this.addHeaders(w)

	// New client channels
	messageChan := make(chan []byte)
	this.newClients <- messageChan

	// Handle client-side disconnection
	ctx := r.Context()

	go func() {
		<-ctx.Done()
		fmt.Println("Client closed connection")
		this.removeClient(messageChan)
	}()

	return messageChan


}

func (this *SSEStreamer) streamLoop(w *http.ResponseWriter, messageChan chan []byte) {

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
		if _, err := fmt.Fprintf(*w, "%s", <-messageChan); err != nil {
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

func (this *SSEStreamer) publish(respData JSONResponseData) {

	body, err := createStreamData(respData)
	if err != nil {
		fmt.Printf("cant get stream data: %s\n", err)
	}

	body = "event:" + getEventName(&respData) + "\ndata:" + body + "\n\n"

	this.Notifier <- []byte(body)

}

func (this *SSEStreamer) addHeaders(writer *http.ResponseWriter) {
	(*writer).Header().Set("Content-Type", "text/event-stream")
	(*writer).Header().Set("Cache-Control", "no-cache")
	(*writer).Header().Set("Connection", "keep-alive")

}

func (this *SSEStreamer) removeClient(msgChan chan []byte) {
	this.closingClients <- msgChan
}

func createStreamData(jsonObj JSONResponseData) (string, error) {

	js, err := json.Marshal(jsonObj)
	if err != nil {
		return "", err
	}
	str := string(js) + "\n\n"
	return str, nil
}

func getEventName(obj *JSONResponseData) string {
	if _, ok := (*obj).(gameStatusResponse); ok {
		return "gamecreated"
	}

	if _, ok := (*obj).(startGameResponse); ok {
		return "gamestarted"
	}

	if _, ok := (*obj).(gameRestartResponse); ok {
		return "gamerestarted"
	}


	if _, ok := (*obj).(gameUpdateResponse); ok {
		return "gameupdated"
	}

	if _, ok := (*obj).(turnUpdateResponse); ok {
		return "gameupdated"
	}

	return ""
}
