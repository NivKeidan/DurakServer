package server

import (
	"DurakGo/game"
	"DurakGo/server/httpPayloadTypes"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)
var (
	invalidCardCodes = []string {"", "%", "!", "?", "QQ", "Qd", "aS", "1S", "000010H", "25H", "\\.", "Ac", "AR", "00006S"}
	invalidPlayerNames = []string{"", "?", "!", "%", "~", "/", "\\", "/\\", "%??~", "\n"}
	invalidPlayerNums = []int{0, 1, 5, -3}
)

func TestCreateGame(t *testing.T) {
	if err := checkMethodsNotAllowed("/createGame", "POST", createGame); err != nil {
		t.Error(err)
	}

	validPlayerName := "niv"
	expectedCode := 400

	for _, invalidPlayerNum := range invalidPlayerNums {
		err := helperCreateGame(invalidPlayerNum, validPlayerName, true, expectedCode)
		if err != nil {
			t.Fatalf("Error: %s\nInvalid player num: %d\n", err.Error(), invalidPlayerNum)
		}
	}

	validPlayerNum := 3
	for _, invalidPlayerName := range invalidPlayerNames {
		err := helperCreateGame(validPlayerNum, invalidPlayerName, true, expectedCode)
		if err != nil {
			t.Fatalf("Error: %s\nInvalid name used: %s\n", err.Error(), invalidPlayerName)
		}
	}

	// Test creating more than one game
	err := helperCreateGame(3, "player1", false, 200)
	if err != nil {
		t.Fatalf("Error: %s\nTest creating several games. First game error\n", err.Error())
	}

	err = helperCreateGame(3, "player2", true, 400)
	if err != nil {
		t.Fatalf("Error: %s\nTest creating several games. Second game error\n", err.Error())
	}

	// Test valid creation with different amount of players
	err = helperCreateGame(2, "niv", true, 200); if err != nil {
		t.Fatalf(err.Error())
	}

	err = helperCreateGame(4, "niv", true, 200); if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestJoinGame(t *testing.T) {
	if err := checkMethodsNotAllowed("/joinGame", "POST", joinGame); err != nil {
		t.Fatal(err)
	}

	validName := "niv"
	expectedCode := 400

	if err := helperJoinGame(validName, expectedCode); err != nil {
		t.Fatalf("Error ocurred while trying to join when game not crated\n" +
			"Error: %s\n", err.Error())
	}

	// Create game
	if err := helperCreateGame(2, "genericCreatorName", false, 200);
		err != nil {
		t.Fatalf("Error ocurred when trrying to create game\n" +
			"Error:: %s\n", err.Error())
	}

	for _, invalidPlayerName := range invalidPlayerNames {
		if err := helperJoinGame(invalidPlayerName, expectedCode); err != nil {
			unCreateGame()
			t.Fatalf("Error ocurred when trying to join with invalid name\n" +
				"Name used: %s\nError: %s\n", invalidPlayerName, err.Error())
		}
	}

	unCreateGame()

	// Create game with 2 players
	name := "testniv"

	if err := helperCreateGame(2, name, false, 200); err != nil {
		t.Fatalf("Could not create game with 2 players. Error: %s\n", err.Error())
	}
	defer unCreateGame()

	// Test join player with same name used for creation
	if err := helperJoinGame(name, 400); err != nil {
		t.Fatalf("Error while testing for joining with same name. Error: %s\n", err.Error())
	}

	// Join second player properly
	if err := helperJoinGame("testniv3", 200); err != nil {
		t.Fatalf("Could not join another player to game. Error: %s\n", err.Error())
	}

	// Test try joining a running game
	if err := helperJoinGame("newName4", 400); err != nil {
		t.Fatalf("Error while testing for joining a running game. Error: %s\n", err.Error())
	}

}

func TestLeaveGame(t *testing.T) {

	if err := checkMethodsNotAllowed("/leaveGame", "POST", leaveGame); err != nil {
		t.Fatal(err)
	}

	// Test leaving when no game created
	validPlayerName := "niv"
	expectedCode := 400

	err := helperLeaveGame(validPlayerName, expectedCode); if err != nil {
		t.Errorf("Error ocurred when testing for leaving without game started\n" +
			"Error:: %s", err.Error())
	}

	// Create game
	if err = helperCreateGame(2, validPlayerName, false, 200); err != nil {
		t.Fatalf("could not create game. Error: %s\n", err.Error())
	}

	defer unCreateGame()

	// Testing leaving with a valid name but that does not exist
	validPlayerName = "niv2"
	if err = helperLeaveGame(validPlayerName, expectedCode); err != nil {
		t.Fatalf("could not leave game with un existing player name\n" +
			"Name used: %s\nError: %s\n", validPlayerName, err.Error())
	}

	// Invalid names
	for _, invalidPlayerName := range invalidPlayerNames {
		if err = helperLeaveGame(invalidPlayerName, expectedCode); err != nil {
			t.Errorf("Error ocurred when testing for leaving game with invalid name\n" +
				"Name used: %s\nError: %s\n", invalidPlayerName, err.Error())
		}
	}

	// Start game
	validPlayerName = "niv2"
	if err := helperJoinGame(validPlayerName, 200); err != nil {
		t.Fatalf("Error ocurred when trying to join (and start) game\n" +
			"Error: %s\n", err.Error())
	}

	// Test leaving when game is running
	if err = helperLeaveGame(validPlayerName, expectedCode); err != nil {
		t.Errorf("Error ocurred when testing for leaving game while game running\n" +
			"Name used: %s\nError: %s\n", validPlayerName, err.Error())
	}
}

func TestAttack(t *testing.T) {
	if err := checkMethodsNotAllowed("/attack", "POST", attack); err != nil {
		t.Error(err)
	}

	// Attacking with no game created
	if err := helperAttack("niv", false, "6D", 400); err != nil {
		t.Fatalf("Error: %s\nTest case: Attacking with no game created\n", err.Error())
	}

	// Create game
	if err := helperCreateGame(2, "niv", false, 200); err != nil {
		t.Fatalf("Could not create game. Error: %s\n", err.Error())
	}

	defer unCreateGame()

	// Attacking with no game started
	if err := helperAttack("niv", false, "6D", 400); err != nil {
		t.Fatalf("Error: %s\nTest case: Attacking with no game started\n", err.Error())
	}

	// Start game
	if err := helperJoinGame("niv2", 200); err != nil {
		t.Fatalf("Could not join (and start) game. Error: %s\n", err.Error())
	}

	// Attacking with invalid card codes
	for _, invalidCardCode := range invalidCardCodes {
		if err := helperAttack("niv", false, invalidCardCode, 400); err != nil {
			t.Fatalf("Error: %s\nTest case: Attacking with invalid card code: %s\n", err.Error(), invalidCardCode)
		}
	}

	// Attacking with invalid player name
	for _, invalidPlayerName := range invalidPlayerNames {
		if err := helperAttack(invalidPlayerName, false, "6D", 400); err != nil {
			t.Fatalf("Error: %s\nTest case: Attacking with invalid player name: %s\n", err.Error(), invalidPlayerName)
		}
	}

	// Attacking with non existing player name
	name := "niv3"
	if err := helperAttack(name, false, "6D", 400); err != nil {
		t.Fatalf("Error: %s\nTest case: Attacking with invalid player name: %s\n", err.Error(), name)
	}

	// Attack regularly
	startingPlayer := currentGame.GetStartingPlayer()
	name = startingPlayer.Name
	if cardCode, err := game.CardToCode(startingPlayer.PeekCards()[0]); err != nil {
		t.Fatalf("Error occurred while trying to get card code from starting player. Error: %s\n", err.Error())
	} else {
		if err := helperAttack(currentGame.GetStartingPlayer().Name, false, cardCode, 200);
			err != nil {
			t.Fatalf("Error: %s\nTest case: Attacking normally: %s\n", err.Error(), name)
		}
	}
}

func TestDefend(t *testing.T) {

	if err := checkMethodsNotAllowed("/defend", "POST", defend); err != nil {
		t.Error(err)
	}

	// Defending with no game created
	if err := helperDefend("niv", "6D", "7D", 400); err != nil {
		t.Fatalf("Error: %s\nTest case: Defending with no game created\n", err.Error())
	}

	// Create game
	if err := helperCreateGame(2, "niv", false, 200); err != nil {
		t.Fatalf("Could not create game. Error: %s\n", err.Error())
	}

	defer unCreateGame()

	// Defending with no game started
	if err := helperDefend("niv", "6D", "7D", 400); err != nil {
		t.Fatalf("Error: %s\nTest case: Defending with no game started\n", err.Error())
	}

	// Start game
	if err := helperJoinGame("niv2", 200); err != nil {
		t.Fatalf("Could not join (and start) game. Error: %s\n", err.Error())
	}

	// Attack with card
	startingPlayer := currentGame.GetStartingPlayer()
	if attCardCode, err := game.CardToCode(startingPlayer.PeekCards()[0]); err != nil {
		t.Fatalf("Could not get card code from player.Error: %s\n", err.Error())
	} else {
		if err := helperAttack(startingPlayer.Name, false, attCardCode, 200); err != nil {
			t.Fatalf("Could not attack.Error: %s\n", err.Error())
		}
	}

	// Defending with invalid attacking card code
	for _, invalidCardCode := range invalidCardCodes {
		if err := helperDefend("niv", invalidCardCode, "7D", 400); err != nil {
			t.Fatalf("Error: %s\nTest case: Defending with no game started\n", err.Error())
		}
	}



	// Defending with invalid defending card code
	// Defending with invalid player name (bad chars, attacking player, non existing, nil)
}

func TestTakeCards(t *testing.T) {

	// Game not created
	// Game not started
	// invalid name
	// non existing name

}

func TestMoveCardsToBita(t *testing.T) {

	// Game not created
	// Game not started
	// invalid name
	// non existing name

}

func TestRestartGame(t *testing.T) {

	// game not created
	// game not started
	// player name not valid
	// player name not existing

}

// Internal non-testing methods

func checkMethodsNotAllowed(endpoint string, methodAllowed string, fn func(w http.ResponseWriter, r *http.Request)) error {
	methods := []string{"GET", "PUT", "POST", "DELETE"}
	for _, method := range methods {
		if method == methodAllowed {
			continue
		}
		req, err := http.NewRequest(method, endpoint, nil)
		if err != nil {
			return err
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(fn)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusMethodNotAllowed {
			return fmt.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusMethodNotAllowed)
		}
	}
	return nil
}

func helperCreateGame(playerNum int, name string, shouldUncreate bool, expectedCode int) error {
	body := httpPayloadTypes.CreateGameRequestObject{
		NumOfPlayers: playerNum,
		PlayerName:   name,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", "/createGame", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("Error occurred: %s\n", err.Error())
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == 200 { // Check for proper response
		resp := httpPayloadTypes.PlayerJoinedResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return err
		}
		if resp.PlayerName != name {
			unCreateGame()
			return fmt.Errorf("Expected returned name to be %s, instead got %s\n", name, resp.PlayerName)
		}
	} else {
		resp := httpPayloadTypes.ErrorResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return err
		}
	}

	if shouldUncreate {
		unCreateGame()
	}

	return nil
}

func helperJoinGame(name string, expectedCode int) error {

	body := httpPayloadTypes.JoinGameRequestObject{
		PlayerName: name,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", "/joinGame", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("Error occurred: %s\n", err.Error())
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(joinGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Join game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == 200 { // Check for proper response
		resp := httpPayloadTypes.PlayerJoinedResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return err
		}
		if resp.PlayerName != name {
			unCreateGame()
			return fmt.Errorf("Expected returned name to be %s, instead got %s\n", name, resp.PlayerName)
		}
	} else {
		resp := httpPayloadTypes.ErrorResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return err
		}
	}

	return nil
}

func helperCreateGameAndJoin(playerNum int, playerNames []string, shouldUncreate bool) error {
	// Create game
	err := helperCreateGame(playerNum, playerNames[0], false, 200); if err != nil {
		return err
	}

	for i := 1; i < len(playerNames); i++ {
		if err := helperJoinGame(playerNames[i], 200); err != nil {
			unCreateGame()
			return err
		}
	}

	if shouldUncreate {
		unCreateGame()
	}

	return nil

}

func helperAttack(name string, shouldUncreateGame bool, cardCode string, expectedCode int) error {
	body := httpPayloadTypes.AttackRequestObject{
		AttackingPlayerName: name,
		AttackingCardCode:   cardCode,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", "/attack", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("Error occurred: %s\n", err.Error())
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(attack)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == 200 { // Check for proper response
		resp := httpPayloadTypes.SuccessResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return err
		}
	} else {
		resp := httpPayloadTypes.ErrorResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			return err
		}
	}

	if shouldUncreateGame {
		unCreateGame()
	}

	return nil
}

func helperLeaveGame(name string, expectedCode int, ) error {
	body := httpPayloadTypes.LeaveGameRequestObject{
		PlayerName: name,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "/leaveGame", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(leaveGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Leave game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == 200 { // Check for proper response
		resp := httpPayloadTypes.SuccessResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			return err
		}
	} else {
		resp := httpPayloadTypes.ErrorResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			return err
		}
	}
	return nil
}

func helperDefend(defendingPlayerName string, attCardCode string, defCardCode string, expectedCode int) error {

	body := httpPayloadTypes.DefenseRequestObject{
		DefendingPlayerName: 	defendingPlayerName,
		DefendingCardCode:   	defCardCode,
		AttackingCardCode:		attCardCode,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", "/defend", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("Error occurred: %s\n", err.Error())
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(defend)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == 200 { // Check for proper response
		resp := httpPayloadTypes.SuccessResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return err
		}
	} else {
		resp := httpPayloadTypes.ErrorResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			return err
		}
	}

	return nil
}