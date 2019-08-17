package stream

import (
	"DurakGo/server/httpPayloadTypes"
	"net/http"
	"time"
)

type AppStreamer struct {
	streamer SSEStreamer
}

func NewAppStreamer(isAliveResp httpPayloadTypes.JSONResponseData) (appStreamer *AppStreamer) {
	appStreamer = &AppStreamer{
		streamer: *NewSSEStreamer(),
	}

	go func() {
		for {
			appStreamer.Publish(isAliveResp)
			time.Sleep(5 * time.Second)
		}
	}()

	return appStreamer
}

func (this *AppStreamer) RegisterClient(w *http.ResponseWriter) chan httpPayloadTypes.JSONResponseData {
	return this.streamer.RegisterClient(w)
}

func (this *AppStreamer) StreamLoop(w *http.ResponseWriter, messageChan chan httpPayloadTypes.JSONResponseData,
	r *http.Request) {

	this.streamer.StreamLoop(w, messageChan, r)
}

func (this *AppStreamer) Publish(respData httpPayloadTypes.JSONResponseData) {

	this.streamer.Publish(respData)
}
