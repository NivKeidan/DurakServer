package stream

import (
	"DurakGo/server"
	"encoding/json"
	"fmt"
)

func convertToString(respData server.JSONResponseData) string {

	body, err := createStreamData(respData)
	if err != nil {
		fmt.Printf("cant get stream data: %s\n", err)
	}

	body = "event:" + getEventName(respData) + "\ndata:" + body + "\n\n"

	return body
}

func createStreamData(jsonObj server.JSONResponseData) (string, error) {

	js, err := json.Marshal(jsonObj)
	if err != nil {
		return "", err
	}
	str := string(js) + "\n\n"
	return str, nil
}

func getEventName(obj server.JSONResponseData) string {
	if _, ok := obj.(*server.gameStatusResponse); ok {
		return "gamestatus"
	}

	if _, ok := obj.(*server.startGameResponse); ok {
		return "gamestarted"
	}

	if _, ok := obj.(*server.gameRestartResponse); ok {
		return "gamerestarted"
	}


	if _, ok := obj.(*server.gameUpdateResponse); ok {
		return "gameupdated"
	}

	if _, ok := obj.(*server.turnUpdateResponse); ok {
		return "gameupdated"
	}

	return ""
}
