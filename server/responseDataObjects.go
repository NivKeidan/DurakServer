package server

import "DurakGo/game"

type JSONResponseData interface {}

type gameStatusResponse struct {
	IsGameRunning bool `json:"isGameRunning"`
	IsGameCreated bool `json:"isGameCreated"`
}

type turnUpdateResponse struct {
	PlayerCards map[string][]*game.Card `json:"playerCards"`
	CardsOnTable []*game.CardOnBoard `json:"cardsOnTable"`
}

type gameUpdateResponse struct {
	PlayerCards          map[string][]*game.Card `json:"playerCards"`
	CardsOnTable         []*game.CardOnBoard     `json:"cardsOnTable"`
	NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
	PlayerStartingName   string                  `json:"playerStarting"`
	PlayerDefendingName  string                  `json:"playerDefending"`
	GameOver             bool                    `json:"gameOver"`
	IsDraw               bool                    `json:"isDraw"`
	LosingPlayerName     string                  `json:"losingPlayerName"`
}

type startGameResponse struct {
	PlayerCards          map[string][]*game.Card `json:"playerCards"`
	KozerCard            *game.Card              `json:"kozerCard"`
	NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
	PlayerStartingName   string                  `json:"playerStarting"`
	PlayerDefendingName  string                  `json:"playerDefending"`
	CardsOnTable         []*game.CardOnBoard     `json:"cardsOnTable"`
	Players				 []string				 `json:"players"`
}

type gameRestartResponse struct {
	PlayerCards          map[string][]*game.Card `json:"playerCards"`
	KozerCard            *game.Card              `json:"kozerCard"`
	CardsOnTable         []*game.CardOnBoard     `json:"cardsOnTable"`
	NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
	PlayerStartingName   string                  `json:"playerStarting"`
	PlayerDefendingName  string                  `json:"playerDefending"`
	GameOver             bool                    `json:"gameOver"`
	IsDraw               bool                    `json:"isDraw"`
}

type playerJoinedResponse struct {
	PlayerName string `json:"playerName"`
	IdCode string `json:"idCode"`
}

// Customize ability for single player clients

type CustomizableJSONResponseData interface {
	GetPlayerCards() map[string][]*game.Card
	SetPlayerCards(m *map[string][]*game.Card)
}

func (this *turnUpdateResponse) GetPlayerCards() map[string][]*game.Card {
	return this.PlayerCards
}

func (this *turnUpdateResponse) SetPlayerCards(m *map[string][]*game.Card)  {
	this.PlayerCards = *m
}

func (this *gameUpdateResponse) GetPlayerCards() map[string][]*game.Card {
	return this.PlayerCards
}

func (this *gameUpdateResponse) SetPlayerCards(m *map[string][]*game.Card)  {
	this.PlayerCards = *m
}

func (this *startGameResponse) GetPlayerCards() map[string][]*game.Card {
	return this.PlayerCards
}

func (this *startGameResponse) SetPlayerCards(m *map[string][]*game.Card)  {
	this.PlayerCards = *m
}

func (this *gameRestartResponse) GetPlayerCards() map[string][]*game.Card {
	return this.PlayerCards
}

func (this *gameRestartResponse) SetPlayerCards(m *map[string][]*game.Card)  {
	this.PlayerCards = *m
}