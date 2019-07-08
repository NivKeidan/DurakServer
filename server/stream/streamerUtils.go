package stream

import (
	"DurakGo/server/httpPayloadTypes"
	"encoding/json"
	"fmt"
)

func convertToString(respData httpPayloadTypes.JSONResponseData) string {

	body, err := createStreamData(respData)
	if err != nil {
		fmt.Printf("cant get stream data: %s\n", err)
	}

	body = "event:" + getEventName(respData) + "\ndata:" + body + "\n\n"

	return body
}

func createStreamData(jsonObj httpPayloadTypes.JSONResponseData) (string, error) {

	js, err := json.Marshal(jsonObj)
	if err != nil {
		return "", err
	}
	str := string(js) + "\n\n"
	return str, nil
}

func getEventName(obj httpPayloadTypes.JSONResponseData) string {
	if _, ok := obj.(*httpPayloadTypes.GameStatusResponse); ok {
		return "gamestatus"
	}

	if _, ok := obj.(*httpPayloadTypes.StartGameResponse); ok {
		return "gamestarted"
	}

	if _, ok := obj.(*httpPayloadTypes.GameRestartResponse); ok {
		return "gamerestarted"
	}


	if _, ok := obj.(*httpPayloadTypes.GameUpdateResponse); ok {
		return "gameupdated"
	}

	if _, ok := obj.(*httpPayloadTypes.TurnUpdateResponse); ok {
		return "gameupdated"
	}

	return ""
}
