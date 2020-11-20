package stream

import (
	"DurakGo/output"
	"DurakGo/server/httpPayloadTypes"
	"time"
)

type AppStreamer struct {
	SSEStreamer
}

func NewAppStreamer(isAliveResp httpPayloadTypes.JSONResponseData, ttl int) (appStreamer *AppStreamer) {
	output.Spit("App streamer running")
	appStreamer = &AppStreamer{
		*NewSSEStreamer(),
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

func (this *AppStreamer) RemoveClient(msgChan chan httpPayloadTypes.JSONResponseData) {
	this.removeClient(msgChan)
}