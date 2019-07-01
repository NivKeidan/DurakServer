package server

import "net/http"

type AppStreamer struct {
	streamer SSEStreamer
}

func NewAppStreamer() (appStreamer *AppStreamer) {
	appStreamer = &AppStreamer{
		streamer: *NewSSEStreamer(),
	}

	return appStreamer
}

func (this *AppStreamer) RegisterClient(w *http.ResponseWriter, r *http.Request) chan JSONResponseData {

	return this.streamer.RegisterClient(w, r)
}

func (this *AppStreamer) StreamLoop(w *http.ResponseWriter, messageChan chan JSONResponseData) {

	this.streamer.StreamLoop(w, messageChan)
}

func (this *AppStreamer) Publish(respData JSONResponseData) {

	this.streamer.Publish(respData)

}
