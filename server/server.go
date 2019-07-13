package server

import (
	"CheekyCommons/stringutil"
	"DurakGo/config"
	"DurakGo/game"
	"DurakGo/server/httpPayloadTypes"
	"DurakGo/server/stream"
	"encoding/json"
	"errors"
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
var configuration *config.Configuration


// External API

func InitServer(conf *config.Configuration) {
	configuration = conf
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
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Extract ID and player name from URL
	keys, ok := r.URL.Query()["id"]
	if !ok {
		http.Error(w, createErrorJson("could not get unique identifier"), http.StatusBadRequest)
		return
	}
	key := keys[0]

	names, ok := r.URL.Query()["name"]
	if !ok {
		http.Error(w, createErrorJson("could not get unique identifier"), http.StatusBadRequest)
		return
	}
	playerName := names[0]

	// Validate player name exists in players

	if !stringutil.IsStringInSlice(playerNames, playerName) {
		http.Error(w, createErrorJson("player name does not exist"), http.StatusBadRequest)
		return
	}

	// Check that ID exists
	if err := validateClientIdentification(playerName, key); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
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
	requestData := httpPayloadTypes.CreateGameRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Create game

	numOfPlayers = requestData.NumOfPlayers
	playerName := requestData.PlayerName

	if err := validateCreateGame(requestData); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	if isGameCreated {
		http.Error(w, createErrorJson("game has already been created"), http.StatusBadRequest)
		return
	}
	handleCreateGame()
	appStreamer.Publish(getGameStatusResponse())

	// Join game

	if err := validateJoinGame(playerName); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		unCreateGame()
		return
	}

	uniquePlayerCode := createPlayerIdentificationString()

	if err := handlePlayerJoin(playerName, uniquePlayerCode); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		unCreateGame()
		return
	}

	if isGameStarted {
		gameStreamer.Publish(getStartGameResponse())
	}

	// Handle response

	if err := integrateJSONResponse(getPlayerJoinedResponse(playerName, uniquePlayerCode), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		unCreateGame()
		return
	}
}

func joinGame(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	requestData := httpPayloadTypes.JoinGameRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	playerName := requestData.PlayerName

	// Validations

	if err := validateJoinGame(playerName); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	uniquePlayerCode := createPlayerIdentificationString()

	if err := handlePlayerJoin(playerName, uniquePlayerCode); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}

	appStreamer.Publish(getGameStatusResponse())
	if isGameStarted {
		gameStreamer.Publish(getStartGameResponse())
	}

	// Handle response
	if err := integrateJSONResponse(getPlayerJoinedResponse(playerName, uniquePlayerCode), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}

}

func leaveGame(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	requestData := httpPayloadTypes.LeaveGameRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	playerName := requestData.PlayerName

	// Validations

	if !isGameCreated {
		http.Error(w, createErrorJson("Create a game first"), http.StatusBadRequest)
		return
	}

	if !stringutil.IsStringInSlice(playerNames, playerName) {
		http.Error(w, createErrorJson("Could not find player"), http.StatusBadRequest)
		return
	}

	if isGameStarted {
		http.Error(w, createErrorJson("game already started"), http.StatusBadRequest)
		return
	}

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
		return
	}

}

func attack(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request


	requestData := httpPayloadTypes.AttackRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Validations
	if !isGameCreated {
		http.Error(w, createErrorJson("game has not been created"), http.StatusBadRequest)
		return
	}

	if !isGameStarted {
		http.Error(w, createErrorJson("game has not been started"), http.StatusBadRequest)
		return
	}


	// Update game
	attackingCard, err := game.NewCardByCode(requestData.AttackingCardCode)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	attackingPlayer, err := currentGame.GetPlayerByName(requestData.AttackingPlayerName)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	if err = currentGame.Attack(attackingPlayer, attackingCard); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Handle response

	gameStreamer.Publish(getUpdateTurnResponse())

	if err = integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}
}

func defend(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request
	requestData := httpPayloadTypes.DefenseRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Validations
	if !isGameCreated {
		http.Error(w, createErrorJson("game has not been created"), http.StatusBadRequest)
		return
	}

	if !isGameStarted {
		http.Error(w, createErrorJson("game has not been started"), http.StatusBadRequest)
		return
	}

	// Update game
	attackingCard, err := game.NewCardByCode(requestData.AttackingCardCode)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	defendingCard, err := game.NewCardByCode(requestData.DefendingCardCode)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	defendingPlayer, err := currentGame.GetPlayerByName(requestData.DefendingPlayerName)

	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	if err = currentGame.Defend(defendingPlayer, attackingCard, defendingCard); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Handle response

	gameStreamer.Publish(getUpdateTurnResponse())

	if err = integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}
}

func takeCards(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request
	requestData := httpPayloadTypes.TakeCardsRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Validations

	if !isGameCreated {
		http.Error(w, createErrorJson("game has not been created"), http.StatusBadRequest)
		return
	}

	if !isGameStarted {
		http.Error(w, createErrorJson("game has not started"), http.StatusBadRequest)
		return
	}

	requestingPlayer, err := currentGame.GetPlayerByName(requestData.PlayerName)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Update game
	if err := currentGame.PickUpCards(requestingPlayer); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Handle response

	gameStreamer.Publish(getUpdateGameResponse())

	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}
}

func moveCardsToBita(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequest(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request
	// Parse request
	requestData := httpPayloadTypes.MoveCardsToBitaObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}


	if !isGameCreated {
		http.Error(w, createErrorJson("game has not been created"), http.StatusBadRequest)
		return
	}

	if !isGameStarted {
		http.Error(w, createErrorJson("game has not started"), http.StatusBadRequest)
		return
	}

	requestingPlayer, err := currentGame.GetPlayerByName(requestData.PlayerName)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Update game
	if err := currentGame.MoveToBita(requestingPlayer); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Handle response

	gameStreamer.Publish(getUpdateGameResponse())

	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
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

	// TODO Validate that this is coming from one of the players
	if !isGameStarted {
		http.Error(w, createErrorJson("no game running at the moment"), http.StatusBadRequest)
		return
	}

	if !currentGame.IsGameOver() {
		http.Error(w, createErrorJson("game is not over"), http.StatusBadRequest)
		return
	}

	// Update game
	if err := startGame(); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}

	// Handle response
	gameStreamer.Publish(getGameRestartResponse())

	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
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
		http.Error(*w, createErrorJson("Method not allowed"), http.StatusMethodNotAllowed)
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

func validateCreateGame(requestData httpPayloadTypes.CreateGameRequestObject) error {
	if requestData.NumOfPlayers < 2 || requestData.NumOfPlayers > 4 {
		return errors.New("can not start game with less than 2 players or more than four players")
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

func getUpdateGameResponse() httpPayloadTypes.JSONResponseData {
	resp := &httpPayloadTypes.GameUpdateResponse{
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

func getUpdateTurnResponse() httpPayloadTypes.JSONResponseData {
	resp := &httpPayloadTypes.TurnUpdateResponse{
		PlayerCards: currentGame.GetPlayersCardsMap(),
		CardsOnTable: currentGame.GetCardsOnBoard(),
	}

	return resp
}

func getStartGameResponse() httpPayloadTypes.JSONResponseData {
	resp := &httpPayloadTypes.StartGameResponse{
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

func getGameStatusResponse() httpPayloadTypes.JSONResponseData {
	resp := &httpPayloadTypes.GameStatusResponse{
		IsGameRunning: isGameStarted,
		IsGameCreated: isGameCreated,
	}

	return resp
}

func getGameRestartResponse() httpPayloadTypes.JSONResponseData {
	resp := &httpPayloadTypes.GameRestartResponse{
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

func getPlayerJoinedResponse(playerName string, code string) httpPayloadTypes.JSONResponseData {
	resp := &httpPayloadTypes.PlayerJoinedResponse{
		PlayerName: playerName,
		IdCode: code,
	}

	return resp
}

// Request/Response Related

func extractJSONData(t httpPayloadTypes.JSONRequestPayload, r *http.Request) error {
	// First argument is the object the data is extracted from
	if err := json.NewDecoder(r.Body).Decode(t); err != nil {
		return err
	}
	return nil
}

func createSuccessJson() httpPayloadTypes.JSONResponseData {
	// Default HTTP JSON body error response

	resp := httpPayloadTypes.SuccessResponse{Message: "", Success: true}
	return resp
}

func integrateJSONResponse(resp httpPayloadTypes.JSONResponseData, w *http.ResponseWriter) error {
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

	resp := httpPayloadTypes.ErrorResponse{Message: errorMessage, Success: false}
	js, _ := json.Marshal(resp)
	return string(js)

}

func addCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", configuration.Get("CorsOrigin"))
	w.Header().Set("Access-Control-Allow-Headers", configuration.Get("CorsHeaders"))
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

func unCreateGame() {
	playerNames = make([]string, 0)
	clientIdentification = make(map[string]map[string]bool)
	isGameCreated = false
	if isGameStarted {
		currentGame = nil
		isGameStarted = false
	}

}

func startGame() error {

	newGame, err := game.NewGame(playerNames...)

	if err != nil {
		return err
	}
	currentGame = newGame
	isGameStarted = true
	return nil
}

func handlePlayerJoin(playerName string, playerUniqueCode string) error {
	// Add player
	if len(playerNames) < numOfPlayers {
		playerNames = append(playerNames, playerName)
	}

	clientIdentification[playerName] = make(map[string]bool)
	clientIdentification[playerName][playerUniqueCode] = false

	// Start game if required
	if len(playerNames) == numOfPlayers {
		if err := startGame(); err != nil {
			return err
		}
	}
	return nil
}

func handleCreateGame() {
	playerNames = make([]string, 0)
	clientIdentification = make(map[string]map[string]bool)
	isGameCreated = true
}

func getCustomizedPlayerCards(respData httpPayloadTypes.CustomizableJSONResponseData,
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

func helperFunc(originalObj httpPayloadTypes.CustomizableJSONResponseData,
	copiedObj httpPayloadTypes.CustomizableJSONResponseData,
	playerName string) error {

	fakePlayerCards := getCustomizedPlayerCards(originalObj, playerName)
	if err := deepcopy.Copy(copiedObj, originalObj); err != nil {
		return errors.New("Error occurred: " + err.Error())
	}
	copiedObj.SetPlayerCards(fakePlayerCards)
	return nil
}

func customizeDataPerPlayer(playerName string) func(httpPayloadTypes.JSONResponseData) (httpPayloadTypes.JSONResponseData, error) {

	return func(respData httpPayloadTypes.JSONResponseData) (httpPayloadTypes.JSONResponseData, error) {
		switch val := respData.(type) {
		case *httpPayloadTypes.StartGameResponse:
			copiedObj := &httpPayloadTypes.StartGameResponse{}
			if err := helperFunc(val, copiedObj, playerName); err != nil {
				return nil, err
			}
			return copiedObj, nil
		case *httpPayloadTypes.GameRestartResponse:
			copiedObj := &httpPayloadTypes.GameRestartResponse{}
			if err := helperFunc(val, copiedObj, playerName); err != nil {
				return nil, err
			}
			return copiedObj, nil
		case *httpPayloadTypes.GameUpdateResponse:
			copiedObj := &httpPayloadTypes.GameUpdateResponse{}
			if err := helperFunc(val, copiedObj, playerName); err != nil {
				return nil, err
			}
			return copiedObj, nil
		case *httpPayloadTypes.TurnUpdateResponse:
			copiedObj := &httpPayloadTypes.TurnUpdateResponse{}
			if err := helperFunc(val, copiedObj, playerName); err != nil {
				return nil, err
			}
			return copiedObj, nil
		default:
			return respData, nil
		}
	}
}
