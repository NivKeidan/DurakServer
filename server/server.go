package server

import (
	"CheekyCommons/stringutil"
	"DurakGo/game"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

var playerNames []string
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
	http.HandleFunc("/restartGame", restartGame)

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
	streamer.publish(getGameStatusResponse())
	if isGameStarted {
		streamer.publish(getStartGameResponse())
	}
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

	// Create game

	numOfPlayers = requestData.NumOfPlayers
	playerName := requestData.PlayerName

	if err := validateCreateGame(requestData); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	handleCreateGame()
	streamer.publish(getGameStatusResponse())

	// Join game

	if err := validateJoinGame(playerName); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	handlePlayerJoin(playerName)

	if isGameStarted {
		streamer.publish(getStartGameResponse())
	}

	// Handle response

	if err := integrateJSONResponse(getPlayerJoinedResponse(playerName), &w); err != nil {
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

	if err := validateJoinGame(playerName); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	handlePlayerJoin(playerName)

	streamer.publish(getGameStatusResponse())
	if isGameStarted {
		streamer.publish(getStartGameResponse())
	}

	// Handle response
	if err := integrateJSONResponse(getPlayerJoinedResponse(playerName), &w); err != nil {
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

	if !stringutil.IsStringInSlice(playerNames, playerName) {
		http.Error(w, createErrorJson("Could not find player"), 400)
		return
	}

	// TODO Handle game already started

	// Remove player
	if len(playerNames) < numOfPlayers {
		playerNames = stringutil.RemoveStringFromSlice(playerNames, playerName)
	}

	// Un-create game if required
	if len(playerNames) == 0 {
		fmt.Printf("Game was cancelled\n")
		isGameCreated = false
	}

	streamer.publish(getGameStatusResponse())

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

	streamer.publish(getUpdateTurnResponse())

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

	streamer.publish(getUpdateTurnResponse())

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

	streamer.publish(getUpdateGameResponse())

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

	streamer.publish(getUpdateGameResponse())

	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func restartGame(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	// Validations

	// Update game
	startGame()

	// Handle response
	streamer.publish(getGameRestartResponse())

	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

// Validations

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

func validateJoinGame(playerName string) error {

	if !isGameCreated {
		return errors.New("create a game first")
	}

	if isGameStarted {
		return errors.New("game has already started")
	}

	if !isNameValid(playerName) {
		return errors.New("player name contains illegal characters")
	}

	if stringutil.IsStringInSlice(playerNames, playerName) {
		return errors.New("name already exists")
	}
	return nil
}

func validateCreateGame(requestData createGameRequestObject) error {
	if requestData.NumOfPlayers < 2 || requestData.NumOfPlayers > 4 {
		return errors.New("can not start game with less than 2 playerNames or more than four playerNames")
	}
	return nil
}

// SSE

func getUpdateGameResponse() JSONResponseData {
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

func getUpdateTurnResponse() JSONResponseData {
	resp := turnUpdateResponse{
		PlayerCards: currentGame.GetPlayersCardsMap(),
		CardsOnTable: currentGame.GetCardsOnBoard(),
	}

	return resp
}

func getStartGameResponse() JSONResponseData {
	resp := startGameResponse {
		PlayerCards: currentGame.GetPlayersCardsMap(),
		KozerCard: currentGame.KozerCard,
		NumOfCardsLeftInDeck: currentGame.GetNumOfCardsLeftInDeck(),
		PlayerStartingName: currentGame.GetStartingPlayer().Name,
		PlayerDefendingName: currentGame.GetDefendingPlayer().Name,
		CardsOnTable:         currentGame.GetCardsOnBoard(),
		Players:			currentGame.GetPlayerNamesArray(),
	}

	return resp
}

func getGameStatusResponse() JSONResponseData {
	resp := gameStatusResponse {
		IsGameRunning: isGameStarted,
		IsGameCreated: isGameCreated,
	}

	return resp
}

func getGameRestartResponse() JSONResponseData {
	resp := gameRestartResponse{
		PlayerCards:          currentGame.GetPlayersCardsMap(),
		KozerCard:            currentGame.KozerCard,
		NumOfCardsLeftInDeck: currentGame.GetNumOfCardsLeftInDeck(),
		PlayerStartingName:   currentGame.GetStartingPlayer().Name,
		PlayerDefendingName:  currentGame.GetDefendingPlayer().Name,
		CardsOnTable:         currentGame.GetCardsOnBoard(),
		GameOver:             currentGame.IsGameOver(),
		IsDraw:				  currentGame.IsDraw(),
	}

	return resp
}

func getPlayerJoinedResponse(playerName string) JSONResponseData {
	resp := playerJoinedResponse{
		PlayerName: playerName,
		IdCode: createPlayerIdentificationString(),
	}

	return resp
}

// Request/Response Related

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

func createPlayerIdentificationString() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*~"
	const length = 10
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}


// Game Logic

func startGame() {

	newGame, err := game.NewGame(playerNames...)

	if err != nil {
		// TODO Handle error
		return
	}
	currentGame = newGame
	isGameStarted = true
}

func handlePlayerJoin(playerName string) {
	// Add player
	if len(playerNames) < numOfPlayers {
		playerNames = append(playerNames, playerName)
	}

	// Start game if required
	if len(playerNames) == numOfPlayers {
		startGame()
	}
}

func handleCreateGame() {
	playerNames = make([]string, 0)
	isGameCreated = true
}