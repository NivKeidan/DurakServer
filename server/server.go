package server

import (
	"CheekyCommons/stringutil"
	"DurakGo/game"
	"DurakGo/server/httpPayloadObjects"
	"DurakGo/server/stream"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getlantern/deepcopy"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

var playerNames []string
var currentGame *game.Game
var isGameCreated = false
var isGameStarted = false
var numOfPlayers int
var appStreamer = stream.NewAppStreamer()
var gameStreamer = stream.NewGameStreamer()
var clientIdentification map[string]map[string]bool


// External API

func InitServer() {
	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/appStream", registerToAppStream)
	http.HandleFunc("/gameStream", registerToGameStream)
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

func registerToAppStream(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"GET"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Register client to appStreamer
	outgoingChannel := appStreamer.RegisterClient(&w, r)
	appStreamer.Publish(getGameStatusResponse())
	if isGameStarted {
		appStreamer.Publish(getStartGameResponse())
	}
	appStreamer.StreamLoop(&w, outgoingChannel)
}

func registerToGameStream(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"GET"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Extract ID and player name from URL
	keys, ok := r.URL.Query()["id"]
	if !ok {
		http.Error(w, createErrorJson("could not get unique identifier"), 400)
		return
	}
	key := keys[0]

	names, ok := r.URL.Query()["name"]
	if !ok {
		http.Error(w, createErrorJson("could not get unique identifier"), 400)
		return
	}
	playerName := names[0]

	// Validate player name exists in players

	if !stringutil.IsStringInSlice(playerNames, playerName) {
		http.Error(w, createErrorJson("player name does not exist"), 400)
		return
	}

	// Check that ID exists
	if err := validateClientIdentification(playerName, key); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	// Open stream and create connection to player

	// Register client to appStreamer
	outgoingChannel := gameStreamer.RegisterClient(&w, r)
	if isGameStarted {
		gameStreamer.Publish(getStartGameResponse())
	}
	gameStreamer.StreamLoop(&w, outgoingChannel, customizeDataPerPlayer(playerName))
}

func createGame(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request
	requestData := httpPayloadObjects.CreateGameRequestObject{}
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
	appStreamer.Publish(getGameStatusResponse())

	// Join game

	if err := validateJoinGame(playerName); err != nil {
		http.Error(w, createErrorJson(err.Error()), 400)
		return
	}

	uniquePlayerCode := createPlayerIdentificationString()

	handlePlayerJoin(playerName, uniquePlayerCode)

	if isGameStarted {
		gameStreamer.Publish(getStartGameResponse())
	}

	// Handle response

	if err := integrateJSONResponse(getPlayerJoinedResponse(playerName, uniquePlayerCode), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

func joinGame(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	requestData := httpPayloadObjects.JoinGameRequestObject{}
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

	uniquePlayerCode := createPlayerIdentificationString()

	handlePlayerJoin(playerName, uniquePlayerCode)

	appStreamer.Publish(getGameStatusResponse())
	if isGameStarted {
		gameStreamer.Publish(getStartGameResponse())
	}

	// Handle response
	if err := integrateJSONResponse(getPlayerJoinedResponse(playerName, uniquePlayerCode), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}

}

func leaveGame(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	requestData := httpPayloadObjects.LeaveGameRequestObject{}
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
		isGameCreated = false
	}

	appStreamer.Publish(getGameStatusResponse())
	gameStreamer.Publish(getGameStatusResponse())

	// Handle response
	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}

}

func attack(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request


	requestData := httpPayloadObjects.AttackRequestObject{}
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

	gameStreamer.Publish(getUpdateTurnResponse())

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


	requestData := httpPayloadObjects.DefenseRequestObject{}
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

	gameStreamer.Publish(getUpdateTurnResponse())

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

	gameStreamer.Publish(getUpdateGameResponse())

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

	gameStreamer.Publish(getUpdateGameResponse())

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
	gameStreamer.Publish(getGameRestartResponse())

	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
	}
}

// Validations

func isNameValid(name string) bool {
	var IsLetter = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString
	return IsLetter(name)

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

func validateCreateGame(requestData httpPayloadObjects.CreateGameRequestObject) error {
	if requestData.NumOfPlayers < 2 || requestData.NumOfPlayers > 4 {
		return errors.New("can not start game with less than 2 playerNames or more than four playerNames")
	}
	return nil
}

func validateClientIdentification(playerName string, code string) error {

	v, ok := clientIdentification[playerName]
	if !ok {
		return errors.New("no such player name registered")
	}

	v2, ok := v[code]
	if !ok {
		return errors.New("identification string is incorrect")
		// TODO Add disconnecting client? (usually means trying to hack or something wrong occurred)
	}
	if v2 {
		return errors.New("client already registered to stream")
		// TODO Add disconnecting client? (usually means trying to hack or something wrong occurred)
	}

	clientIdentification[playerName][code] = true
	return nil
}

// SSE

func getUpdateGameResponse() httpPayloadObjects.JSONResponseData {
	resp := &httpPayloadObjects.GameUpdateResponse{
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

func getUpdateTurnResponse() httpPayloadObjects.JSONResponseData {
	resp := &httpPayloadObjects.TurnUpdateResponse{
		PlayerCards: currentGame.GetPlayersCardsMap(),
		CardsOnTable: currentGame.GetCardsOnBoard(),
	}

	return resp
}

func getStartGameResponse() httpPayloadObjects.JSONResponseData {
	resp := &httpPayloadObjects.StartGameResponse{
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

func getGameStatusResponse() httpPayloadObjects.JSONResponseData {
	resp := &httpPayloadObjects.GameStatusResponse{
		IsGameRunning: isGameStarted,
		IsGameCreated: isGameCreated,
	}

	return resp
}

func getGameRestartResponse() httpPayloadObjects.JSONResponseData {
	resp := &httpPayloadObjects.GameRestartResponse{
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

func getPlayerJoinedResponse(playerName string, code string) httpPayloadObjects.JSONResponseData {
	resp := &httpPayloadObjects.PlayerJoinedResponse{
		PlayerName: playerName,
		IdCode: code,
	}

	return resp
}

// Request/Response Related

func extractJSONData(t httpPayloadObjects.JSONRequestPayload, r *http.Request) error {
	// First argument is the object the data is extracted from
	if err := json.NewDecoder(r.Body).Decode(t); err != nil {
		return err
	}
	return nil
}

func createSuccessJson() httpPayloadObjects.JSONResponseData {
	// Default HTTP JSON body error response

	type errorResponse struct {
		Success bool `json:"success"`
		Message string `json:"message"`
	}

	resp := errorResponse{Message: "", Success: true}
	return resp
}

func integrateJSONResponse(resp httpPayloadObjects.JSONResponseData, w *http.ResponseWriter) error {
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
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
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

func handlePlayerJoin(playerName string, playerUniqueCode string) {
	// Add player
	if len(playerNames) < numOfPlayers {
		playerNames = append(playerNames, playerName)
	}

	clientIdentification[playerName] = make(map[string]bool)
	clientIdentification[playerName][playerUniqueCode] = false

	// Start game if required
	if len(playerNames) == numOfPlayers {
		startGame()
	}
}

func handleCreateGame() {
	playerNames = make([]string, 0)
	clientIdentification = make(map[string]map[string]bool)
	isGameCreated = true
}

func getCustomizedPlayerCards(respData httpPayloadObjects.CustomizableJSONResponseData,
	playerName string) *map[string][]*game.Card {

	fakePlayerCards := make(map[string][]*game.Card)

	for k, v := range respData.GetPlayerCards() {
		if k != playerName {
			fakePlayerCards[k] = make([]*game.Card, len(v))
		} else {
			fakePlayerCards[k] = v
		}
	}
	return &fakePlayerCards
}

func helperFunc(originalObj httpPayloadObjects.CustomizableJSONResponseData,
	copiedObj httpPayloadObjects.CustomizableJSONResponseData,
	playerName string) httpPayloadObjects.CustomizableJSONResponseData {

	fakePlayerCards := getCustomizedPlayerCards(originalObj, playerName)
	if err := deepcopy.Copy(copiedObj, originalObj); err != nil {
		// TODO Handle this error better?
		fmt.Printf("ERROR OCCURRED: %v\n", err)
		return nil
	}
	copiedObj.SetPlayerCards(fakePlayerCards)
	return copiedObj
}

func customizeDataPerPlayer(playerName string) func(httpPayloadObjects.JSONResponseData) httpPayloadObjects.JSONResponseData {

	return func(respData httpPayloadObjects.JSONResponseData) httpPayloadObjects.JSONResponseData {
		switch val := respData.(type) {
		case *httpPayloadObjects.StartGameResponse:
			copiedObj := &httpPayloadObjects.StartGameResponse{}
			return helperFunc(val, copiedObj, playerName)
		case *httpPayloadObjects.GameRestartResponse:
			copiedObj := &httpPayloadObjects.GameRestartResponse{}
			return helperFunc(val, copiedObj, playerName)
		case *httpPayloadObjects.GameUpdateResponse:
			copiedObj := &httpPayloadObjects.GameUpdateResponse{}
			return helperFunc(val, copiedObj, playerName)
		case *httpPayloadObjects.TurnUpdateResponse:
			copiedObj := &httpPayloadObjects.TurnUpdateResponse{}
			return helperFunc(val, copiedObj, playerName)
		default:
			return respData
		}
	}
}
