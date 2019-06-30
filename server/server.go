package server

import (
	"CheekyCommons/stringutil"
	"DurakGo/game"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

var players []string
var currentGame *game.Game
var isGameCreated = false
var isGameStarted = false
var numOfPlayers int
var streamer = NewStreamer()

// External API

func InitServer() {
	http.HandleFunc("/eventSource", registerToStream)
	http.HandleFunc("/createGame", createGame)
	http.HandleFunc("/joinGame", joinGame)
	http.HandleFunc("/leaveGame", leaveGame)
	http.HandleFunc("/attack", attack)
	http.HandleFunc("/defend", defend)
	http.HandleFunc("/takeCards", takeCards)
	http.HandleFunc("/moveCardsToBita", moveCardsToBita)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// API

func registerToStream(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"GET"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Register client to streamer
	outgoingChannel := streamer.registerClient(&w, r)
	streamer.publish(getGameStatus())
	streamer.streamLoop(&w, outgoingChannel)
}

func createGame(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request
	requestData := createGameRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
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
	players = make([]string, 0)
	isGameCreated = true

	// Handle response
	streamer.publish(getGameStatus())

	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
	fmt.Println("New game created")
}

func joinGame(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	requestData := joinGameRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	playerName := requestData.PlayerName

	// Validations

	if !isGameCreated {
		http.Error(w, createErrorJson("Create a game first"), 400)
		return
	}

	if isGameStarted {
		http.Error(w, createErrorJson("Game has already started"), 400)
		return
	}

	if !isNameValid(requestData.PlayerName) {
		http.Error(w, createErrorJson("Player name contains illegal characters"), 400)
		return
	}

	if stringutil.IsStringInSlice(players, playerName) {
		http.Error(w, createErrorJson("Name already exists"), 400)
		return
	}

	// Add player
	if len(players) < numOfPlayers {
		players = append(players, playerName)
	}

	// Start game if required
	if len(players) == numOfPlayers {
		initializeGame()
		isGameStarted = true
	}

	streamer.publish(getGameStatus())
	if isGameStarted {
		streamer.publish(startGame())
	}

	// Handle response
	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}

	fmt.Printf("%v has joined the game\n", playerName)
}

func leaveGame(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	requestData := leaveGameRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	playerName := requestData.PlayerName

	// Validations

	if !isGameCreated {
		http.Error(w, createErrorJson("Create a game first"), 400)
		return
	}

	if !stringutil.IsStringInSlice(players, playerName) {
		http.Error(w, createErrorJson("Could not find player"), 400)
		return
	}

	// TODO Handle game already started

	// Remove player
	if len(players) < numOfPlayers {
		players = stringutil.RemoveStringFromSlice(players, playerName)
	}

	// Un-create game if required
	if len(players) == 0 {
		fmt.Printf("Game was cancelled\n")
		isGameCreated = false
	}

	streamer.publish(getGameStatus())

	// Handle response
	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}

	fmt.Printf("%v has left the game\n", playerName)
}

func isNameValid(name string) bool {
	// TODO Add name validations
	return true

}

func attack(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request


	requestData := attackRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
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

	if err = currentGame.Attack(attackingPlayer, attackingCard); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Handle response

	streamer.publish(updateTurn())

	if err = integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func defend(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request


	requestData := defenseRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
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

	if err = currentGame.Defend(defendingPlayer, attackingCard, defendingCard); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Handle response

	streamer.publish(updateTurn())

	if err = integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func takeCards(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	// Validations

	// Update game
	currentGame.HandlePlayerTakesCard()

	// Handle response

	streamer.publish(updateGame())

	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func moveCardsToBita(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	// Validations

	// Update game
	if err := currentGame.MoveToBita(); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Handle response

	streamer.publish(updateGame())

	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

// SSE

func updateGame() JSONResponseData {
	resp := gameUpdateResponse{
		PlayerCards:          currentGame.GetPlayersCardsMap(),
		CardsOnTable:         currentGame.GetCardsOnBoard(),
		NumOfCardsLeftInDeck: currentGame.GetNumOfCardsLeftInDeck(),
		PlayerStartingName:   currentGame.GetStartingPlayer().Name,
		PlayerDefendingName:  currentGame.GetDefendingPlayer().Name,
		GameOver:             currentGame.IsGameOver(),
		IsDraw:				  currentGame.IsDraw(),
		LosingPlayerName:	  currentGame.GetLosingPlayerName(),
	}

	return resp
}

func updateTurn() JSONResponseData {
	resp := turnUpdateResponse{
		PlayerCards: currentGame.GetPlayersCardsMap(),
		CardsOnTable: currentGame.GetCardsOnBoard(),
	}

	return resp
}

func startGame() JSONResponseData {
	resp := startGameResponse {
		PlayerCards: currentGame.GetPlayersCardsMap(),
		KozerCard: currentGame.KozerCard,
		NumOfCardsLeftInDeck: currentGame.GetNumOfCardsLeftInDeck(),
		PlayerStartingName: currentGame.GetStartingPlayer().Name,
		PlayerDefendingName: currentGame.GetDefendingPlayer().Name,
	}

	return resp
}

func getGameStatus() JSONResponseData {
	resp := gameStatusResponse {
		IsGameRunning: isGameStarted,
		IsGameCreated: isGameCreated,
	}

	return resp
}

// Internal methods

func initializeGame() {

	newGame, err := game.NewGame(players...)

	if err != nil {
		// TODO Handle error
		return
	}
	currentGame = newGame
}

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
	if err := json.NewDecoder(r.Body).Decode(t); err != nil {
		return err
	}
	return nil
}

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
	if _, err = (*w).Write(js); err != nil {
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