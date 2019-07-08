package stream

import (
	"DurakGo/server/httpPayloadTypes"
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

func (this *AppStreamer) RegisterClient(w *http.ResponseWriter, r *http.Request) chan httpPayloadTypes.JSONResponseData {

	return this.streamer.RegisterClient(w, r)
}

func (this *AppStreamer) StreamLoop(w *http.ResponseWriter, messageChan chan httpPayloadTypes.JSONResponseData) {

	this.streamer.StreamLoop(w, messageChan)
}

func (this *AppStreamer) Publish(respData httpPayloadTypes.JSONResponseData) {

	this.streamer.Publish(respData)

}
