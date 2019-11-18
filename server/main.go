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

	output.Spit("Server initialized!")

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
	http.HandleFunc("/connectionId", createConnectionId)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Game Logic

func unCreateGame() {
	output.Spit("Uncreating game")
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
	output.Spit(fmt.Sprintf("Player %s joined", user))
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

func handleCreateGame() {
	output.Spit("Creating game")
	//users = make([]*User, 0)
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
