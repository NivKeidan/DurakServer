package server

import (
	"sync"
)

type GameManager struct {
	games []*GameHolder
	gameCreatorLock	*sync.Mutex
	currentOpenGame *GameHolder
	lastIdUsed int
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make([]*GameHolder, 0),
		gameCreatorLock: &sync.Mutex{},
		currentOpenGame: nil,
		lastIdUsed: 0,
	}
}

func (this *GameManager) CreateNewGame(playerNum int) *GameHolder {
	this.gameCreatorLock.Lock()
	defer func() { this.gameCreatorLock.Unlock() }()

	if this.currentOpenGame != nil {
		return nil
	} else {
		this.lastIdUsed++
		gameHolder := NewGameHolder(this.lastIdUsed, playerNum)
		this.currentOpenGame = gameHolder
		this.games = append(this.games, this.currentOpenGame)
	}
}

func (this *GameManager) IsGameCreated() bool {
	this.gameCreatorLock.Lock()
	defer func() { this.gameCreatorLock.Unlock() }()

	return this.currentOpenGame != nil
}