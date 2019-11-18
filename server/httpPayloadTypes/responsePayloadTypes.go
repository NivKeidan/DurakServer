package httpPayloadTypes

import "DurakGo/game"

type JSONResponseData interface {}

type GameStatusResponse struct {
	IsGameRunning bool `json:"isGameRunning"`
	IsGameCreated bool `json:"isGameCreated"`
}

type TurnUpdateResponse struct {
	PlayerCards map[string][]*game.Card `json:"playerCards"`
	CardsOnTable []*game.CardOnBoard `json:"cardsOnTable"`
}

type GameUpdateResponse struct {
	PlayerCards          map[string][]*game.Card `json:"playerCards"`
	CardsOnTable         []*game.CardOnBoard     `json:"cardsOnTable"`
	NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
	PlayerStartingName   string                  `json:"playerStarting"`
	PlayerDefendingName  string                  `json:"playerDefending"`
	GameOver             bool                    `json:"gameOver"`
	IsDraw               bool                    `json:"isDraw"`
	LosingPlayerName     string                  `json:"losingPlayerName"`
}

type StartGameResponse struct {
	PlayerCards          map[string][]*game.Card `json:"playerCards"`
	KozerCard            *game.Card              `json:"kozerCard"`
	NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
	PlayerStartingName   string                  `json:"playerStarting"`
	PlayerDefendingName  string                  `json:"playerDefending"`
	CardsOnTable         []*game.CardOnBoard     `json:"cardsOnTable"`
	Players				 []string				 `json:"players"`
}

type GameRestartResponse struct {
	PlayerCards          map[string][]*game.Card `json:"playerCards"`
	KozerCard            *game.Card              `json:"kozerCard"`
	CardsOnTable         []*game.CardOnBoard     `json:"cardsOnTable"`
	NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
	PlayerStartingName   string                  `json:"playerStarting"`
	PlayerDefendingName  string                  `json:"playerDefending"`
	GameOver             bool                    `json:"gameOver"`
	IsDraw               bool                    `json:"isDraw"`
}

type PlayerJoinedResponse struct {
	PlayerName string `json:"playerName"`
	IdCode string `json:"idCode"`
}

type IsAliveResponse struct {}

type GetConnectionIdResponse struct {
	ConnectionId string `json:"connectionId"`
}

type ErrorResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
}

// Customize ability for single player clients

type CustomizableJSONResponseData interface {
	GetPlayerCards() map[string][]*game.Card
	SetPlayerCards(m *map[string][]*game.Card)
}

func (this *TurnUpdateResponse) GetPlayerCards() map[string][]*game.Card {
	return this.PlayerCards
}

func (this *TurnUpdateResponse) SetPlayerCards(m *map[string][]*game.Card)  {
	this.PlayerCards = *m
}

func (this *GameUpdateResponse) GetPlayerCards() map[string][]*game.Card {
	return this.PlayerCards
}

func (this *GameUpdateResponse) SetPlayerCards(m *map[string][]*game.Card)  {
	this.PlayerCards = *m
}

func (this *StartGameResponse) GetPlayerCards() map[string][]*game.Card {
	return this.PlayerCards
}

func (this *StartGameResponse) SetPlayerCards(m *map[string][]*game.Card)  {
	this.PlayerCards = *m
}

func (this *GameRestartResponse) GetPlayerCards() map[string][]*game.Card {
	return this.PlayerCards
}

func (this *GameRestartResponse) SetPlayerCards(m *map[string][]*game.Card)  {
	this.PlayerCards = *m
}