package httpPayloadTypes

type JSONRequestPayload interface {}

type CreateGameRequestObject struct {
	NumOfPlayers int `json:"numOfPlayers"`
	PlayerName string `json:"playerName"`
}

type JoinGameRequestObject struct {
	PlayerName string `json:"playerName"`
}

type LeaveGameRequestObject struct {
	PlayerName string `json:"playerName"`
}

type AttackRequestObject struct {
	AttackingPlayerName string `json:"attackingPlayerName"`
	AttackingCardCode string `json:"attackingCardCode"`
}

type DefenseRequestObject struct {
	DefendingPlayerName string `json:"defendingPlayerName"`
	DefendingCardCode string `json:"defendingCardCode"`
	AttackingCardCode string `json:"attackingCardCode"`
}