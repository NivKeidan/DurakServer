package server

import (
	"DurakGo/server/httpPayloadTypes"
	"math/rand"
)

type User struct {
	connectionId string
	gameChan chan httpPayloadTypes.JSONResponseData
	appChan chan httpPayloadTypes.JSONResponseData
	name string
	isAlive bool
	isJoined bool
}

func NewUser(name string) *User {
	return &User{connectionId: createPlayerIdentificationString(), name: name, isAlive: true, isJoined: false,
		gameChan:nil, appChan: nil}
}

func createPlayerIdentificationString() string {
	letters := configuration.GetString("ClientIdLetters")
	length := configuration.GetInt("ClientIdLength")
	b := make([]byte, length)
	var s string
	for doesCodeExist(s) {
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		s = string(b)
	}
	return s
}