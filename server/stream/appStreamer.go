package stream

import (
	"DurakGo/server"
	"net/http"
)

type AppStreamer struct {
	streamer SSEStreamer
}

func NewAppStreamer() (appStreamer *AppStreamer) {
	appStreamer = &AppStreamer{
		streamer: *NewSSEStreamer(),
	}

	return appStreamer
}

func (this *AppStreamer) RegisterClient(w *http.ResponseWriter, r *http.Request) chan server.JSONResponseData {

	return RegisterClient(w, r)
}

func (this *AppStreamer) StreamLoop(w *http.ResponseWriter, messageChan chan server.JSONResponseData) {

	StreamLoop(w, messageChan)
}

func (this *AppStreamer) Publish(respData server.JSONResponseData) {

	Publish(respData)

}
