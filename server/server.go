package server

import (
	"CheekyCommons/stringutil"
	"DurakGo/game"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

var players = make([]string, 0)
var currentGame *game.Game
var isGameCreated = false
var isGameStarted = false
var numOfPlayers int

// External API

func InitServer() {
	http.HandleFunc("/createGame", createGame)
	http.HandleFunc("/joinGame", joinGame)
	http.HandleFunc("/attack", attack)
	http.HandleFunc("/defend", defend)
	http.HandleFunc("/takeCards", takeCards)
	http.HandleFunc("/moveCardsToBita", moveCardsToBita)
	http.HandleFunc("/gameStatus", getCurrentGameStatus)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getCurrentGameStatus(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"GET"}
	err := validateRequest(&w, r, allowedMethods)
	if err != nil {
		return
	}

	// Handle response

	type gameStatusResponse struct {
		IsGameRunning bool `json:"isGameRunning"`
		IsGameCreated bool `json:"isGameCreated"`
	}

	resp := gameStatusResponse{
		IsGameRunning: isGameStarted,
		IsGameCreated: isGameCreated,
	}

	err = integrateJSONResponse(&resp, &w)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func createGame(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	err := validateRequest(&w, r, allowedMethods)
	if err != nil {
		return
	}

	// Parse request
	type optionsObject struct {
		NumOfPlayers int `json:"numOfPlayers"`
	}

	requestData := optionsObject{}
	err = extractJSONData(&requestData, r)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Validations

	if requestData.NumOfPlayers < 2 || requestData.NumOfPlayers > 4 {
		http.Error(w, "Can not start game with less than 2 players or more than four players", 400)
		return
	}

	// Initializations

	numOfPlayers = requestData.NumOfPlayers
	players = make([]string, numOfPlayers)
	isGameCreated = true

	// Handle response
	err = integrateJSONResponse(createSuccessJson(), &w)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func joinGame(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	err := validateRequest(&w, r, allowedMethods)
	if err != nil {
		return
	}

	// Parse request
	type optionsObject struct {
		PlayerName string `json:"playerName"`
	}

	requestData := optionsObject{}
	err = extractJSONData(&requestData, r)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Validations

	if !isGameCreated {
		http.Error(w, createErrorJson("Create a game first"), 400)
	}

	if isGameStarted {
		http.Error(w, createErrorJson("Game has already started"), 400)
	}

	if !isNameValid(requestData.PlayerName) {
		http.Error(w, createErrorJson("Player name contains illegal characters"), 400)
	}

	if stringutil.IsStringInSlice(players, requestData.PlayerName) {
		http.Error(w, createErrorJson("Name already exists"), 400)
	}

	// Add player
	if len(players) < numOfPlayers {
		players = append(players, requestData.PlayerName)
	}

	// Start game if required
	if len(players) == numOfPlayers {
		initializeGame()
		isGameStarted = true
		eventSourceDummyGameStarted()
	} else {
		eventSourceDummyPlayerJoined()
	}

	// Handle response
	err = integrateJSONResponse(createSuccessJson(), &w)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func isNameValid(name string) bool {
	// TODO Add name validations
	return true

}

func initializeGame() {

	newGame, err := game.NewGame(players...)

	if err != nil {
		eventSourceDummyFailToStartGame(err)
		return
	}
	currentGame = newGame
}

func eventSourceDummyPlayerJoined() {
	// Dummy for sending out eventsource event for another player joined
	// but game is still not starting
}

func eventSourceDummyGameStarted() {
	// Dummy for sending out eventsource event for game starting

	//type startGameResponse struct {
	//	PlayerCards          map[string][]*game.Card `json:"playerCards"`
	//	KozerCard            *game.Card              `json:"kozerCard"`
	//	NumOfCardsLeftInDeck int                     `json:"numOfCardsLeftInDeck"`
	//	PlayerStartingName   string                  `json:"playerStarting"`
	//	PlayerDefendingName  string                  `json:"playerDefending"`
	//}
	//
	//resp := startGameResponse {
	//	KozerCard:            currentGame.KozerCard,
	//	NumOfCardsLeftInDeck: currentGame.NumOfCardsLeftInDeck(),
	//	PlayerStartingName:   currentGame.GetStartingPlayer().Name,
	//	PlayerDefendingName:  currentGame.GetDefendingPlayer().Name,
	//	PlayerCards:          currentGame.GetPlayersCardsMap(),
	//}

	//err = integrateJSONResponse(&resp, &w)
	//if err != nil {
	//	http.Error(w, createErrorJson(err.Error()), 500)
	//}
}

func eventSourceDummyFailToStartGame(e error) {
	// Dummy for sending out eventsource event for failure to start game
}

func attack(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	err := validateRequest(&w, r, allowedMethods)
	if err != nil {
		return
	}

	// Parse request
	type attackObject struct {
		AttackingPlayerName string `json:"attackingPlayerName"`
		AttackingCardCode string `json:"attackingCardCode"`
	}

	requestData := attackObject{}
	err = extractJSONData(&requestData, r)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Validations


	// Update game
	attackingCard, err := game.NewCardByCode(requestData.AttackingCardCode)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	attackingPlayer, err := currentGame.GetPlayerByName(requestData.AttackingPlayerName)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	err = currentGame.Attack(attackingPlayer, attackingCard)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Handle response

	type attackResponse struct {
		PlayerCards map[string][]*game.Card `json:"playerCards"`
		CardsOnTable []*game.CardOnBoard `json:"cardsOnTable"`
	}

	resp := attackResponse{
		PlayerCards: currentGame.GetPlayersCardsMap(),
		CardsOnTable: currentGame.GetCardsOnBoard(),
	}

	err = integrateJSONResponse(&resp, &w)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func defend(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	err := validateRequest(&w, r, allowedMethods)
	if err != nil {
		return
	}

	// Parse request
	type defenseObject struct {
		DefendingPlayerName string `json:"defendingPlayerName"`
		DefendingCardCode string `json:"defendingCardCode"`
		AttackingCardCode string `json:"attackingCardCode"`
	}

	requestData := defenseObject{}
	err = extractJSONData(&requestData, r)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Validations

	// Update game
	attackingCard, err := game.NewCardByCode(requestData.AttackingCardCode)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	defendingCard, err := game.NewCardByCode(requestData.DefendingCardCode)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	defendingPlayer, err := currentGame.GetPlayerByName(requestData.DefendingPlayerName)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	err = currentGame.Defend(defendingPlayer, attackingCard, defendingCard)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Handle response

	type defenseResponse struct {
		PlayerCards map[string][]*game.Card `json:"playerCards"`
		CardsOnTable []*game.CardOnBoard `json:"cardsOnTable"`
	}

	resp := defenseResponse{
		PlayerCards: currentGame.GetPlayersCardsMap(),
		CardsOnTable: currentGame.GetCardsOnBoard(),
	}

	err = integrateJSONResponse(&resp, &w)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func takeCards(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	err := validateRequest(&w, r, allowedMethods)
	if err != nil {
		return
	}

	// Parse request

	// Validations

	// Update game
	currentGame.HandlePlayerTakesCard()

	// Handle response
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
		PlayerCards:          currentGame.GetPlayersCardsMap(),
		CardsOnTable:         currentGame.GetCardsOnBoard(),
		NumOfCardsLeftInDeck: currentGame.GetNumOfCardsLeftInDeck(),
		PlayerStartingName:   currentGame.GetStartingPlayer().Name,
		PlayerDefendingName:  currentGame.GetDefendingPlayer().Name,
		GameOver:             currentGame.IsGameOver(),
		IsDraw:				  currentGame.IsDraw(),
		LosingPlayerName:	  currentGame.GetLosingPlayerName(),
	}

	err = integrateJSONResponse(&resp, &w)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func moveCardsToBita(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	err := validateRequest(&w, r, allowedMethods)
	if err != nil {
		return
	}

	// Parse request

	// Validations

	// Update game
	err = currentGame.MoveToBita()

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Handle response

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
		PlayerCards:          currentGame.GetPlayersCardsMap(),
		CardsOnTable:         currentGame.GetCardsOnBoard(),
		NumOfCardsLeftInDeck: currentGame.GetNumOfCardsLeftInDeck(),
		PlayerStartingName:   currentGame.GetStartingPlayer().Name,
		PlayerDefendingName:  currentGame.GetDefendingPlayer().Name,
		GameOver:             currentGame.IsGameOver(),
		IsDraw:				  currentGame.IsDraw(),
		LosingPlayerName:	  currentGame.GetLosingPlayerName(),
	}

	err = integrateJSONResponse(&resp, &w)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

// Internal methods

func validateRequest(w *http.ResponseWriter, r *http.Request, allowedMethods []string) error {
	// Handles CORS, HTTP Method

	// TODO Upgrade CORS handling
	addCorsHeaders(*w)
	if r.Method == "OPTIONS" {
		_, _ = (*w).Write([]byte("OK"))
		return errors.New("send response back now")

	} else if !isMethodAllowed(r, allowedMethods) {
		http.Error(*w, createErrorJson("Method not allowed"), 405)
		return errors.New("send response back now")
	}
	return nil
}

func extractJSONData(t JSONRequestPayload, r *http.Request) error {
	// First argument is the object the data is extracted from
	err := json.NewDecoder(r.Body).Decode(t)
	if err != nil {
		return err
	}
	return nil
}

type JSONRequestPayload interface {}

type JSONResponseData interface {}

func createSuccessJson() JSONResponseData {
	// Default HTTP JSON body error response

	type errorResponse struct {
		Success bool `json:"success"`
		Message string `json:"message"`
	}

	resp := errorResponse{Message: "", Success: true}
	return resp
}

func integrateJSONResponse(resp JSONResponseData, w *http.ResponseWriter) error {
	// First argument is the object the data is put into

	js, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	(*w).Header().Set("Content-Type", "application/json")
	_, err = (*w).Write(js)

	if err != nil {
		return err
	}
	return nil
}

func isMethodAllowed(request *http.Request, allowedMethods []string) bool {
	return stringutil.IsStringInSlice(allowedMethods, request.Method)
}

func createErrorJson(errorMessage string) string {
	// Default HTTP JSON body error response

	type errorResponse struct {
		Success bool `json:"success"`
		Message string `json:"message"`
	}

	resp := errorResponse{Message: errorMessage, Success: false}
	js, _ := json.Marshal(resp)
	return string(js)

}

func addCorsHeaders(w http.ResponseWriter) {
	// TODO Integrate this from config file

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}