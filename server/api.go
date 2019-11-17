package server

import (
	"CheekyCommons/stringutil"
	"DurakGo/game"
	"DurakGo/output"
	"DurakGo/server/httpPayloadTypes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

func registerToAppStream(w http.ResponseWriter, r *http.Request) {

	// Validate request headers
	allowedMethods := []string{"GET"}
	if err := validateRequestMethod(&w, r, allowedMethods); err != nil {
		return
	}

	output.Spit("a user registered to app stream")

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

	output.Spit(fmt.Sprintf("user ID %s registered to game stream", user))

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