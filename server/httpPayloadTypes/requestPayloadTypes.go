package httpPayloadTypes

type JSONRequestPayload interface {}

type CreateGameRequestObject struct {
	NumOfPlayers int `json:"numOfPlayers"`
	PlayerName string `json:"playerName"`
}

type JoinGameRequestObject struct {
	PlayerName string `json:"playerName"`
}

type AttackRequestObject struct {
	AttackingCardCode string `json:"attackingCardCode"`
}

type DefenseRequestObject struct {
	DefendingCardCode string `json:"defendingCardCode"`
	AttackingCardCode string `json:"attackingCardCode"`
}