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