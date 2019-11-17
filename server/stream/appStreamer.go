package stream

import (
	"DurakGo/output"
	"DurakGo/server/httpPayloadTypes"
	"net/http"
	"time"
)

type AppStreamer struct {
	streamer SSEStreamer
}

func NewAppStreamer(isAliveResp httpPayloadTypes.JSONResponseData, ttl int) (appStreamer *AppStreamer) {
	output.Spit("App streamer running")
	appStreamer = &AppStreamer{
		streamer: *NewSSEStreamer(),
	}

	go func() {
		output.Spit("go routine - publish is alive app streamer - start")
		defer func() {
			output.Spit("go routine - publish is alive app streamer - ended")
		}()
		for {
			appStreamer.Publish(isAliveResp)
			time.Sleep(time.Duration(ttl / 2) * time.Second)
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
