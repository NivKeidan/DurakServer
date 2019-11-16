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

var users []*User
var currentGame *game.Game
var isGameCreated bool
var isGameStarted bool
var numOfPlayers int
var appStreamer *stream.AppStreamer
var gameStreamer *stream.GameStreamer
var configuration *config.Configuration
var aliveTTL int
var notAliveUser chan *User

func InitServer(conf *config.Configuration) {

	configuration = conf
	aliveTTL = conf.GetInt("AliveTTL")
	isGameStarted = false
	isGameCreated = false
	appStreamer = stream.NewAppStreamer(getIsAliveResponse(), aliveTTL)
	gameStreamer = stream.NewGameStreamer(getIsAliveResponse(), aliveTTL)
	notAliveUser = make(chan *User)

	go handleDeadUsers()

	rand.Seed(time.Now().UnixNano())
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
	http.HandleFunc("/alive", alive)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// API

func registerToAppStream(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"GET"}
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Register client to appStreamer
	outgoingChannel := appStreamer.RegisterClient(&w)
	appStreamer.Publish(getGameStatusResponse())
	if isGameStarted {
		appStreamer.Publish(getStartGameResponse())
	}
	appStreamer.StreamLoop(&w, outgoingChannel, r)
}

func registerToGameStream(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"GET"}
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Extract ID and player name from URL
	keys, ok := r.URL.Query()["id"]
	if !ok {
		http.Error(w, createErrorJson("could not get unique identifier from URL"), http.StatusBadRequest)
		return
	}
	connectionId := keys[0]

	if !isGameCreated {
		http.Error(w, createErrorJson("Game has not been created yet"), http.StatusBadRequest)
		return
	}

	if !isGameStarted {
		http.Error(w, createErrorJson("Game has not started yet"), http.StatusBadRequest)
		return
	}

	user := getUserByConnectionId(connectionId)
	if user == nil {
		http.Error(w, createErrorJson("Could not find player"), http.StatusBadRequest)
		return
	}

	// Open stream and create connection to player

	outgoingChannel := gameStreamer.RegisterClient(&w)
	user.gameChan = outgoingChannel

	if isGameStarted {
		gameStreamer.Publish(getStartGameResponse())
	}


	gameStreamer.StreamLoop(&w, outgoingChannel, r, customizeDataPerPlayer(user.name))
}

func createGame(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request
	requestData := httpPayloadTypes.CreateGameRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	numOfPlayers = requestData.NumOfPlayers
	playerName := requestData.PlayerName

	// Create game

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

	// Create User

	if err := validatePlayerName(playerName); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		unCreateGame()
		return
	}

	newUser := NewUser(playerName, aliveTTL, notAliveUser)

	// Join game

	if err := validateJoinGame(); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		unCreateGame()
		return
	}

	if err := handlePlayerJoin(newUser); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		unCreateGame()
		return
	}

	if isGameStarted {
		gameStreamer.Publish(getStartGameResponse())
	}

	// Handle response

	if err := integrateJSONResponse(getPlayerJoinedResponse(newUser), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		unCreateGame()
		return
	}
}

func joinGame(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"POST"}
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Parse request

	requestData := httpPayloadTypes.JoinGameRequestObject{}
	if err := extractJSONData(&requestData, r); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	playerName := requestData.PlayerName

	if err := validatePlayerName(playerName); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Validations

	if err := validateJoinGame(); err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	newUser := NewUser(playerName, aliveTTL, notAliveUser)

	if err := handlePlayerJoin(newUser); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}

	appStreamer.Publish(getGameStatusResponse())
	if isGameStarted {
		gameStreamer.Publish(getStartGameResponse())
	}

	// Handle response
	if err := integrateJSONResponse(getPlayerJoinedResponse(newUser), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}

}

func leaveGame(w http.ResponseWriter, r *http.Request) {
	// Validate request method
	allowedMethods := []string{"POST"}
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Validate connection id
	connectionId, err := getConnectionId(r)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Validations

	if !isGameCreated {
		http.Error(w, createErrorJson("Create a game first"), http.StatusBadRequest)
		return
	}

	user := getUserByConnectionId(connectionId)
	if user == nil {
		http.Error(w, createErrorJson("Could not find player"), http.StatusBadRequest)
		return
	}

	if isGameStarted {
		http.Error(w, createErrorJson("game already started"), http.StatusBadRequest)
		return
	}

	// Remove player
	users = removeUser(user)

	// Un-create game if required
	if len(users) == 0 {
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
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Validate connection id
	connectionId, err := getConnectionId(r)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

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

	user := getUserByConnectionId(connectionId)
	if user == nil {
		http.Error(w, createErrorJson("Could not find player"), http.StatusBadRequest)
		return
	}
	user.receivedAlive()

	attackingPlayer, err := currentGame.GetPlayerByName(user.name)

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
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Validate connection id
	connectionId, err := getConnectionId(r)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
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

	user := getUserByConnectionId(connectionId)
	if user == nil {
		http.Error(w, createErrorJson("Could not find player"), http.StatusBadRequest)
		return
	}
	user.receivedAlive()

	defendingPlayer, err := currentGame.GetPlayerByName(user.name)

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
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Validate connection id
	connectionId, err := getConnectionId(r)
	if err != nil {
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

	user := getUserByConnectionId(connectionId)
	if user == nil {
		http.Error(w, createErrorJson("Could not find player"), http.StatusBadRequest)
		return
	}

	user.receivedAlive()
	if !isUserPlaying(user) {
		http.Error(w, createErrorJson("user is not a player"), http.StatusBadRequest)
		return
	}

	// Update game
	if err := currentGame.PickUpCards(); err != nil {
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
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Validate connection id
	connectionId, err := getConnectionId(r)
	if err != nil {
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

	user := getUserByConnectionId(connectionId)
	if user == nil {
		http.Error(w, createErrorJson("Could not find player"), http.StatusBadRequest)
		return
	}

	user.receivedAlive()

	if !isUserPlaying(user) {
		http.Error(w, createErrorJson("user is not a player"), http.StatusBadRequest)
		return
	}

	// Update game
	if err := currentGame.MoveToBita(); err != nil {
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
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Validate connection id
	_, err := getConnectionId(r)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	// Validations

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

func alive(w http.ResponseWriter, r *http.Request) {
	// Validate request headers
	allowedMethods := []string{"GET"}
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	// Validate connection id
	connId, err := getConnectionId(r)
	if err != nil {
		http.Error(w, createErrorJson(err.Error()), http.StatusBadRequest)
		return
	}

	user := getUserByConnectionId(connId)
	if user == nil {
		http.Error(w, createErrorJson("Could not find player"), http.StatusBadRequest)
		return
	}

	user.receivedAlive()

	// Handle response
	if err := integrateJSONResponse(createSuccessJson(), &w); err != nil {
		http.Error(w, createErrorJson(err.Error()), 500)
		return
	}
}

// Validations

func getConnectionId(r *http.Request) (string, error) {
	connId := r.Header.Get("ConnectionId")
	if connId == "" {
		return "", errors.New("connection id header missing")
	}
	return connId, nil
}

func isNameValid(name string) bool {
	var IsLetter = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString
	return IsLetter(name)

}

func validateRequestMethod(w *http.ResponseWriter, r *http.Request, allowedMethods []string) error {
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

func validateJoinGame() error {

	if !isGameCreated {
		return errors.New("create a game first")
	}

	if isGameStarted {
		return errors.New("game has already started")
	}

	return nil
}

func validateCreateGame(requestData httpPayloadTypes.CreateGameRequestObject) error {
	if requestData.NumOfPlayers < 2 || requestData.NumOfPlayers > 4 {
		return errors.New("can not start game with less than 2 players or more than four players")
	}
	return nil
}

func validatePlayerName(name string) error {

	if !isNameValid(name) {
		return errors.New("player name contains illegal characters")
	}

	for _, u := range users {
		if u.name == name {
			return errors.New("name already exists")
		}
	}

	return nil
}

func isUserPlaying(user *User) bool {
	_, err := currentGame.GetPlayerByName(user.name)
	return err == nil
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

func getPlayerJoinedResponse(user *User) httpPayloadTypes.JSONResponseData {
	resp := &httpPayloadTypes.PlayerJoinedResponse{
		PlayerName: user.name,
		IdCode: user.connectionId,
	}

	return resp
}

func getIsAliveResponse() httpPayloadTypes.JSONResponseData {
	resp := &httpPayloadTypes.IsAliveResponse{}
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
	w.Header().Set("Access-Control-Allow-Origin", configuration.GetString("CorsOrigin"))
	w.Header().Set("Access-Control-Allow-Headers", configuration.GetString("CorsHeaders"))
}

func doesCodeExist(c string) bool {
	// This func is called in a loop, so first call should return true
	if c == "" {
		return true
	}

	for _, u := range users {
		if c == u.connectionId {
			return true
		}
	}

	return false
}

// Game Logic

func unCreateGame() {
	users = make([]*User, 0)
	isGameCreated = false
	if isGameStarted {
		currentGame = nil
		isGameStarted = false
	}

}

func startGame() error {

	playerNames := make([]string, 0)

	for _, u := range users {
		playerNames = append(playerNames, u.name)
	}

	newGame, err := game.NewGame(playerNames...)

	if err != nil {
		return err
	}
	currentGame = newGame
	isGameStarted = true
	return nil
}

func handlePlayerJoin(user *User) error {
	if len(users) < numOfPlayers {
		users = append(users, user)
	}

	// Start game if required
	if len(users) == numOfPlayers {
		if err := startGame(); err != nil {
			return err
		}
	}
	return nil
}

func handleCreateGame() {
	users = make([]*User, 0)
	isGameCreated = true
}

func getUserByConnectionId(connId string) *User {
	for _, u := range users {
		if u.connectionId == connId {
			return u
		}
	}
	return nil
}

func removeUser(u *User) []*User {
	for i, user := range users {
		if user == u {
			return append(users[:i], users[i+1:]...)
		}
	}
	return users
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
	copiedObj httpPayloadTypes.CustomizableJSONResponseData, playerName string) error {

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

func handleDeadUsers() {
	for {
		deadUser := <- notAliveUser
		if isGameCreated {
			if ! isGameStarted {
				users = removeUser(deadUser)

				// Un-create game if required
				if len(users) == 0 {
					isGameCreated = false
				}

				appStreamer.Publish(getGameStatusResponse())
			} else {
				playerName := deadUser.name
				if err := currentGame.HandlePlayerLeft(playerName); err != nil {

				}
				gameStreamer.Publish(getUpdateGameResponse())
			}
		}


	}
}
