package server

import (
	"DurakGo/game"
	"DurakGo/server/stream"
)

type GameHolder struct {
	ID int
	users []*User
	game *game.Game
	isGameStarted bool
	numOfPlayers int
	gameStreamer *stream.GameStreamer
}

func NewGameHolder(id int, playerNum int) *GameHolder{
	return &GameHolder{
		ID: id,
		isGameStarted: false,
		numOfPlayers: playerNum,
		gameStreamer: stream.NewGameStreamer()
	}

}
