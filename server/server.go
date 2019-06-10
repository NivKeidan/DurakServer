package server

import (
	"DurakGo/game"
	"encoding/json"
	"log"
	"net/http"
)

var players = make([]string, 0)
var currentGame *game.Game

// External API

func InitServer() {
	http.HandleFunc("/startGame", startGame)
	http.HandleFunc("/attack", attack)
	http.HandleFunc("/defend", defend)
	http.HandleFunc("/takeCards", takeCards)
	http.HandleFunc("/moveCardsToBita", moveCardsToBita)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func startGame(w http.ResponseWriter, r *http.Request) {
	/*
		Initializes board, deals cards, chooses a kozer card, and sets first turn.
	 */

	// CORS
	// TODO Better CORS handling
	addCorsHeaders(w)
	if r.Method == "OPTIONS" {
		_, _ = w.Write([]byte("OK"))
		return
	} else if r.Method != "POST" {
		http.Error(w, createErrorJson("Method not allowed"), 405)
	}

	// Handle request object
	type optionsObject struct {
		NumOfPlayers int `json:"numOfPlayers"`
	}

	var reqBodyObject optionsObject
	err := json.NewDecoder(r.Body).Decode(&reqBodyObject)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Validations

	if reqBodyObject.NumOfPlayers < 2 || reqBodyObject.NumOfPlayers > 4 {
		http.Error(w, "Can not start game with less than 2 players or more than four players", 400)
		return
	}

	// Create objects

	players = make([]string, 0)
	namesArray := []string{"Niv", "Asaf", "Vala", "Roee"}
	for i := 0; i < reqBodyObject.NumOfPlayers; i++ {
		players = append(players, namesArray[i])
	}

	newGame, err := game.NewGame(players...)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}
	currentGame = newGame

	// Handle response

	type startGameResponse struct {
		PlayerCards          map[string][]*game.Card `json:"playerCards"`
		KozerCard            *game.Card              `json:"kozerCard"`
		NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
		PlayerStartingName   string                  `json:"playerStarting"`
		PlayerDefendingName  string                  `json:"playerDefending"`
	}

	resp := startGameResponse {
		KozerCard:            currentGame.KozerCard,
		NumOfCardsLeftInDeck: currentGame.NumOfCardsLeftInDeck(),
		PlayerStartingName:   currentGame.GetStartingPlayer().Name,
		PlayerDefendingName:  currentGame.GetDefendingPlayer().Name,
		PlayerCards:          getPlayerCardsMap(),
	}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}

}

func attack(w http.ResponseWriter, r *http.Request) {

	// CORS

	// TODO Add better CORS handling
	addCorsHeaders(w)
	if r.Method == "OPTIONS" {
		_, _ = w.Write([]byte("OK"))
		return
	} else if r.Method != "POST" {
		http.Error(w, createErrorJson("Method not allowed"), 405)
	}

	// Validations
	//userName := r.Header.Get(USER_HEADER)

	// Parse request body
	type attackObject struct {
		AttackingPlayerName string `json:"attackingPlayerName"`
		AttackingCardCode string `json:"attackingCardCode"`
	}

	var reqBodyObject attackObject
	err := json.NewDecoder(r.Body).Decode(&reqBodyObject)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Perform action
	attackingCard, err := game.NewCardByCode(reqBodyObject.AttackingCardCode)


	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	attackingPlayer, err := currentGame.GetPlayerByName(reqBodyObject.AttackingPlayerName)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	err = currentGame.Attack(attackingPlayer, attackingCard)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	type attackResponse struct {
		PlayerCards map[string][]*game.Card `json:"playerCards"`
		CardsOnTable []*game.CardOnBoard `json:"cardsOnTable"`
	}

	resp := attackResponse{
		PlayerCards: getPlayerCardsMap(),
		CardsOnTable: currentGame.GetCardsOnBoard(),
	}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}
}

func defend(w http.ResponseWriter, r *http.Request) {

	// CORS

	// TODO Add better CORS handling
	addCorsHeaders(w)
	if r.Method == "OPTIONS" {
		_, _ = w.Write([]byte("OK"))
		return
	} else if r.Method != "POST" {
		http.Error(w, createErrorJson("Method not allowed"), 405)
	}

	// Validations
	//userName := r.Header.Get(USER_HEADER)

	// Parse request body
	type defenseObject struct {
		DefendingPlayerName string `json:"defendingPlayerName"`
		DefendingCardCode string `json:"defendingCardCode"`
		AttackingCardCode string `json:"attackingCardCode"`
	}

	var reqBodyObject defenseObject
	err := json.NewDecoder(r.Body).Decode(&reqBodyObject)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Perform action
	attackingCard, err := game.NewCardByCode(reqBodyObject.AttackingCardCode)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	defendingCard, err := game.NewCardByCode(reqBodyObject.DefendingCardCode)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	defendingPlayer, err := currentGame.GetPlayerByName(reqBodyObject.DefendingPlayerName)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	err = currentGame.Defend(defendingPlayer, attackingCard, defendingCard)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	type defenseResponse struct {
		PlayerCards map[string][]*game.Card `json:"playerCards"`
		CardsOnTable []*game.CardOnBoard `json:"cardsOnTable"`
	}

	resp := defenseResponse{
		PlayerCards: getPlayerCardsMap(),
		CardsOnTable: currentGame.GetCardsOnBoard(),
	}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}


}

func takeCards(w http.ResponseWriter, r *http.Request) {
	// CORS
	// TODO Add better CORS handling
	addCorsHeaders(w)
	if r.Method == "OPTIONS" {
		_, _ = w.Write([]byte("OK"))
		return
	} else if r.Method != "POST" {
		http.Error(w, createErrorJson("Method not allowed"), 405)
	}

	// Validations
	//userName := r.Header.Get(USER_HEADER)

	// TODO Add more validations

	// Perform action
	currentGame.HandlePlayerTakesCard()

	type takeCardsResponse struct {
		PlayerCards          map[string][]*game.Card `json:"playerCards"`
		CardsOnTable         []*game.CardOnBoard     `json:"cardsOnTable"`
		NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
		PlayerStartingName   string                  `json:"playerStarting"`
		PlayerDefendingName  string                  `json:"playerDefending"`
		GameOver             bool                    `json:"gameOver"`
		IsDraw               bool                    `json:"isDraw"`
		LosingPlayerName     string                  `json:"losingPlayerName"`
	}

	resp := takeCardsResponse{
		PlayerCards:          getPlayerCardsMap(),
		CardsOnTable:         currentGame.GetCardsOnBoard(),
		NumOfCardsLeftInDeck: currentGame.GetNumOfCardsLeftInDeck(),
		PlayerStartingName:   currentGame.GetStartingPlayer().Name,
		PlayerDefendingName:  currentGame.GetDefendingPlayer().Name,
		GameOver:             currentGame.IsGameOver(),
	}

	if resp.GameOver {
		resp.IsDraw =  currentGame.IsDraw()
		if !resp.IsDraw {
			losingPlayer, err := currentGame.GetLosingPlayer()

			if err != nil {
				http.Error(w, createErrorJson(err.Error()), 400)
				return
			}
			resp.LosingPlayerName = losingPlayer.Name
		}

	} else {
		resp.IsDraw = false
		resp.LosingPlayerName = ""
	}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}
}

func moveCardsToBita(w http.ResponseWriter, r *http.Request) {
	// CORS
	// TODO Add better CORS handling
	addCorsHeaders(w)
	if r.Method == "OPTIONS" {
		_, _ = w.Write([]byte("OK"))
		return
	} else if r.Method != "POST" {
		http.Error(w, createErrorJson("Method not allowed"), 405)
	}
	// Validations
	//userName := r.Header.Get(USER_HEADER)

	// Perform action
	err := currentGame.MoveToBita()

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	type newTurnResponse struct {
		PlayerCards          map[string][]*game.Card `json:"playerCards"`
		CardsOnTable         []*game.CardOnBoard     `json:"cardsOnTable"`
		NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
		PlayerStartingName   string                  `json:"playerStarting"`
		PlayerDefendingName  string                  `json:"playerDefending"`
		GameOver             bool                    `json:"gameOver"`
		IsDraw               bool                    `json:"isDraw"`
		LosingPlayerName     string                  `json:"losingPlayerName"`
	}

	resp := newTurnResponse{
		PlayerCards:          getPlayerCardsMap(),
		CardsOnTable:         currentGame.GetCardsOnBoard(),
		NumOfCardsLeftInDeck: currentGame.GetNumOfCardsLeftInDeck(),
		PlayerStartingName:   currentGame.GetStartingPlayer().Name,
		PlayerDefendingName:  currentGame.GetDefendingPlayer().Name,
		GameOver:             currentGame.IsGameOver(),
	}
	if resp.GameOver {
		resp.IsDraw =  currentGame.IsDraw()
		if !resp.IsDraw {
			losingPlayer, err := currentGame.GetLosingPlayer()
			if err != nil {
				http.Error(w, createErrorJson(err.Error()), 400)
				return
			}
			resp.LosingPlayerName = losingPlayer.Name
		}
	} else {
		resp.IsDraw = false
		resp.LosingPlayerName = ""
	}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}
}

// Internal methods

func createErrorJson(errorMessage string) string {
	type errorResponse struct {
		Message string `json:"message"`
	}

	resp := errorResponse{Message: errorMessage}
	js, _ := json.Marshal(resp)
	return string(js)

}

func addCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func getPlayerCardsMap() map[string][]*game.Card {
	playerCards := make(map[string][]*game.Card)
	for _, playerName := range players {
		player, _ := currentGame.GetPlayerByName(playerName)
		cards := player.GetAllCards()
		playerCards[playerName] = cards
	}

	return playerCards

}