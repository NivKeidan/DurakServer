package server

import (
	"encoding/json"
	"fmt"
)

func convertToString(respData JSONResponseData) string {

	body, err := createStreamData(respData)
	if err != nil {
		fmt.Printf("cant get stream data: %s\n", err)
	}

	body = "event:" + getEventName(&respData) + "\ndata:" + body + "\n\n"

	return body
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
