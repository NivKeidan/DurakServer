package stream

import (
	"DurakGo/server/httpPayloadObjects"
	"encoding/json"
	"fmt"
)

func convertToString(respData httpPayloadObjects.JSONResponseData) string {

	body, err := createStreamData(respData)
	if err != nil {
		fmt.Printf("cant get stream data: %s\n", err)
	}

	body = "event:" + getEventName(respData) + "\ndata:" + body + "\n\n"

	return body
}

func createStreamData(jsonObj httpPayloadObjects.JSONResponseData) (string, error) {

	js, err := json.Marshal(jsonObj)
	if err != nil {
		return "", err
	}
	str := string(js) + "\n\n"
	return str, nil
}

func getEventName(obj httpPayloadObjects.JSONResponseData) string {
	if _, ok := obj.(*httpPayloadObjects.GameStatusResponse); ok {
		return "gamestatus"
	}

	if _, ok := obj.(*httpPayloadObjects.StartGameResponse); ok {
		return "gamestarted"
	}

	if _, ok := obj.(*httpPayloadObjects.GameRestartResponse); ok {
		return "gamerestarted"
	}


	if _, ok := obj.(*httpPayloadObjects.GameUpdateResponse); ok {
		return "gameupdated"
	}

	if _, ok := obj.(*httpPayloadObjects.TurnUpdateResponse); ok {
		return "gameupdated"
	}

	return ""
}
