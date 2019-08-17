package server

import "DurakGo/server/httpPayloadTypes"

type User struct {
	connectionId string
	gameChan chan httpPayloadTypes.JSONResponseData
	name string
	isAlive bool
	isJoined bool
}

func NewUser(connectionId string, name string) *User {
	return &User{connectionId: connectionId, name: name, isAlive: true, isJoined: false, gameChan: nil}
}

