package server

type JSONRequestPayload interface {}

type createGameRequestObject struct {
	NumOfPlayers int `json:"numOfPlayers"`
	PlayerName string `json:"playerName"`
}

type joinGameRequestObject struct {
	PlayerName string `json:"playerName"`
}

type leaveGameRequestObject struct {
	PlayerName string `json:"playerName"`
}

type attackRequestObject struct {
	AttackingPlayerName string `json:"attackingPlayerName"`
	AttackingCardCode string `json:"attackingCardCode"`
}

type defenseRequestObject struct {
	DefendingPlayerName string `json:"defendingPlayerName"`
	DefendingCardCode string `json:"defendingCardCode"`
	AttackingCardCode string `json:"attackingCardCode"`
}