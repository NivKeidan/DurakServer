package server

import (
	"DurakGo/config"
	"DurakGo/game"
	"DurakGo/output"
	"DurakGo/server/httpPayloadTypes"
	"DurakGo/server/stream"
	"errors"
	"fmt"
	"github.com/getlantern/deepcopy"
	"log"
	"math/rand"
	"net/http"
	"time"
)


var gameManager *GameManager
var userManager *UserManager
var appStreamer *stream.AppStreamer
var configuration *config.Configuration

func InitServer(conf *config.Configuration) {

	output.Spit("Server initialized!")
	aliveTTL := conf.GetInt("AliveTTL")
	gameManager = NewGameManager()
	userManager = NewUserManager()
	appStreamer = stream.NewAppStreamer(getIsAliveResponse(), aliveTTL)

	go handleDeadUsers()

	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/connectionId", createConnectionId)
	http.HandleFunc("/appStream", registerToAppStream)
	http.HandleFunc("/alive", alive)
	http.HandleFunc("/createGame", createGame)
	http.HandleFunc("/joinGame", joinGame)
	http.HandleFunc("/gameStream", registerToGameStream)
	http.HandleFunc("/leaveGame", leaveGame)
	http.HandleFunc("/attack", attack)
	http.HandleFunc("/defend", defend)
	http.HandleFunc("/takeCards", takeCards)
	http.HandleFunc("/moveCardsToBita", moveCardsToBita)
	http.HandleFunc("/restartGame", restartGame)


	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Game Logic

func unCreateGame() {
	output.Spit("Uncreating game")
	gameManager.UncreateGame()
	//users = make([]*User, 0)
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

	output.Spit("Starting game!")

	newGame, err := game.NewGame(playerNames...)

	if err != nil {
		return err
	}
	currentGame = newGame
	isGameStarted = true
	return nil
}

func handlePlayerJoin(user *User) error {
	output.Spit(fmt.Sprintf("User %s generated Player %s and joined to game", user.connectionId, user.name))
	user.isJoined = true

	// Start game if required
	if getNumOfJoinedUsers() == numOfPlayers {
		if err := startGame(); err != nil {
			return err
		}
	}
	return nil
}

func getNumOfJoinedUsers() int {
	i := 0
	for _, u := range users {
		if u.isJoined {
			i++
		}
	}
	return i
}

func removeUser(u *User) []*User {
	output.Spit(fmt.Sprintf("Removing user %s", u))
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
	output.Spit("go routine - monitoring dead users - start")
	defer func() {
		output.Spit("go routine - monitoring dead users - ended")
	}()

	for {
		deadUser := <-notAliveUser
		output.Spit(fmt.Sprintf("User %s is dead. Removing from app stream", deadUser))
		appStreamer.RemoveClient(deadUser.appChan)
		if isGameCreated {
			output.Spit(fmt.Sprintf("User %s is dead. Removing from game stream", deadUser))
			appStreamer.RemoveClient(deadUser.gameChan)
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
					// TODO What is missing here???
				}
				gameStreamer.Publish(getUpdateGameResponse())
			}
		}

	}
}
