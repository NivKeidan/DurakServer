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
	invalidClientCode = helperCreateInvalidClientCode()
)

func TestCreateGame(t *testing.T) {
	if err := checkMethodsNotAllowed("/createGame", "POST", createGame); err != nil {
		t.Error(err)
	}

	validPlayerName := "niv"
	expectedCode := http.StatusBadRequest

	for _, invalidPlayerNum := range invalidPlayerNums {
		_, err := helperCreateGame(invalidPlayerNum, validPlayerName, true, expectedCode)
		if err != nil {
			t.Fatalf("Error: %s\nInvalid player num: %d\n", err.Error(), invalidPlayerNum)
		}
	}

	validPlayerNum := 3
	for _, invalidPlayerName := range invalidPlayerNames {
		_, err := helperCreateGame(validPlayerNum, invalidPlayerName, true, expectedCode)
		if err != nil {
			t.Fatalf("Error: %s\nInvalid name used: %s\n", err.Error(), invalidPlayerName)
		}
	}

	// Test creating more than one game
	_, err := helperCreateGame(3, "player1", false, http.StatusOK)
	if err != nil {
		t.Fatalf("Error: %s\nTest creating several games. First game error\n", err.Error())
	}

	_, err = helperCreateGame(3, "player2", true, http.StatusBadRequest)
	if err != nil {
		t.Fatalf("Error: %s\nTest creating several games. Second game error\n", err.Error())
	}

	// Test valid creation with different amount of players
	if _, err = helperCreateGame(2, "niv", true, http.StatusOK); err != nil {
		t.Fatalf(err.Error())
	}

	if _, err = helperCreateGame(4, "niv", true, http.StatusOK); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestJoinGame(t *testing.T) {
	if err := checkMethodsNotAllowed("/joinGame", "POST", joinGame); err != nil {
		t.Fatal(err)
	}

	validName := "niv"
	expectedCode := http.StatusBadRequest

	// Join when no game created
	if _, err := helperJoinGame(validName, expectedCode); err != nil {
		t.Fatalf("Error ocurred while trying to join when game not crated\n" +
			"Error: %s\n", err.Error())
	}

	// Create game
	if _, err := helperCreateGame(2, "genericCreatorName", false, http.StatusOK);
		err != nil {
		t.Fatalf("Error ocurred when trrying to create game\n" +
			"Error:: %s\n", err.Error())
	}

	for _, invalidPlayerName := range invalidPlayerNames {
		if _, err := helperJoinGame(invalidPlayerName, expectedCode); err != nil {
			unCreateGame()
			t.Fatalf("Error ocurred when trying to join with invalid name\n" +
				"Name used: %s\nError: %s\n", invalidPlayerName, err.Error())
		}
	}

	unCreateGame()

	// Create game with 2 players
	name := "testniv"

	if _, err := helperCreateGame(2, name, false, http.StatusOK); err != nil {
		t.Fatalf("Could not create game with 2 players. Error: %s\n", err.Error())
	}
	defer unCreateGame()

	// Test join player with same name used for creation
	if _, err := helperJoinGame(name, http.StatusBadRequest); err != nil {
		t.Fatalf("Error while testing for joining with same name. Error: %s\n", err.Error())
	}

	// Join second player properly
	if _, err := helperJoinGame("testniv3", http.StatusOK); err != nil {
		t.Fatalf("Could not join another player to game. Error: %s\n", err.Error())
	}

	// Test try joining a running game
	if _, err := helperJoinGame("newName4", http.StatusBadRequest); err != nil {
		t.Fatalf("Error while testing for joining a running game. Error: %s\n", err.Error())
	}

}

func TestLeaveGame(t *testing.T) {

	if err := checkMethodsNotAllowed("/leaveGame", "POST", leaveGame); err != nil {
		t.Fatal(err)
	}

	// Test leaving when no game created
	validPlayerName := "niv"
	expectedCode := http.StatusBadRequest

	err := helperLeaveGame(invalidClientCode, expectedCode); if err != nil {
		t.Errorf("Error ocurred when testing for leaving without game started\n" +
			"Error:: %s", err.Error())
	}

	// Create game
	if _, err := helperCreateGame(2, validPlayerName, false, http.StatusOK); err != nil {
		t.Fatalf("could not create game. Error: %s\n", err.Error())
	}

	defer unCreateGame()

	// Testing leaving with a valid name but that does not exist
	if err = helperLeaveGame(invalidClientCode, expectedCode); err != nil {
		t.Fatalf("could not leave game with un existing client code\n" +
			"Client code: %s\nError: %s\n", invalidClientCode, err.Error())
	}

	// Start game
	validPlayerName = "niv2"
	clientCode, err := helperJoinGame(validPlayerName, http.StatusOK)
	if err != nil {
		t.Fatalf("Error ocurred when trying to join (and start) game\n" +
			"Error: %s\n", err.Error())
	}

	// Test leaving when game is running
	if err = helperLeaveGame(clientCode, expectedCode); err != nil {
		t.Errorf("Error ocurred when testing for leaving game while game running\n" +
			"Client code: %s\nError: %s\n", clientCode, err.Error())
	}
}

func TestAttack(t *testing.T) {
	if err := checkMethodsNotAllowed("/attack", "POST", attack); err != nil {
		t.Error(err)
	}

	// Attacking with no game created
	if err := helperAttack(invalidClientCode, false, "6D", http.StatusBadRequest); err != nil {
		t.Fatalf("Error: %s\nTest case: Attacking with no game created\n", err.Error())
	}

	// Create game
	clientCode, err := helperCreateGame(2, "niv", false, http.StatusOK)
	if err != nil {
		t.Fatalf("Could not create game. Error: %s\n", err.Error())
	}

	defer unCreateGame()

	// Attacking with no game started
	if err := helperAttack(clientCode, false, "6D", http.StatusBadRequest); err != nil {
		t.Fatalf("Error: %s\nTest case: Attacking with no game started\n", err.Error())
	}

	// Start game
	if _, err := helperJoinGame("niv2", http.StatusOK); err != nil {
		t.Fatalf("Could not join (and start) game. Error: %s\n", err.Error())
	}

	// Attacking with invalid card codes
	for _, invalidCardCode := range invalidCardCodes {
		if err := helperAttack(clientCode, false, invalidCardCode, http.StatusBadRequest); err != nil {
			t.Fatalf("Error: %s\nTest case: Attacking with invalid card code: %s\n", err.Error(), invalidCardCode)
		}
	}

	// Attacking with non existing client code
	if err := helperAttack(invalidClientCode, false, "6D", http.StatusBadRequest); err != nil {
		t.Fatalf("Error: %s\nTest case: Attacking with invalid client code: %s\n", err.Error(), invalidClientCode)
	}

	// Attack regularly
	startingPlayer := currentGame.GetStartingPlayer()
	name := startingPlayer.Name
	clientCode = helperGetClientCodeByName(name)
	if cardCode, err := game.CardToCode(startingPlayer.PeekCards()[0]); err != nil {
		t.Fatalf("Error occurred while trying to get card code from starting player. Error: %s\n", err.Error())
	} else {
		if err := helperAttack(clientCode, false, cardCode, http.StatusOK);
			err != nil {
			t.Fatalf("Error: %s\nTest case: Attacking normally\n", err.Error())
		}
	}
}

func TestDefend(t *testing.T) {

	if err := checkMethodsNotAllowed("/defend", "POST", defend); err != nil {
		t.Error(err)
	}

	// Defending with no game created
	if err := helperDefend("niv", "6D", "7D", http.StatusBadRequest); err != nil {
		t.Fatalf("Error: %s\nTest case: Defending with no game created\n", err.Error())
	}

	// Create game
	clientCode, err := helperCreateGame(2, "niv", false, http.StatusOK)
	if err != nil {
		t.Fatalf("Could not create game. Error: %s\n", err.Error())
	}

	defer unCreateGame()

	// Defending with no game started
	if err := helperDefend(clientCode, "6D", "7D", http.StatusBadRequest); err != nil {
		t.Fatalf("Error: %s\nTest case: Defending with no game started\n", err.Error())
	}

	// Start game
	if _, err := helperJoinGame("niv2", http.StatusOK); err != nil {
		t.Fatalf("Could not join (and start) game. Error: %s\n", err.Error())
	}

	// Attack with card
	startingPlayer := currentGame.GetStartingPlayer()
	attCard := startingPlayer.PeekCards()[0]
	attCardCode, err := game.CardToCode(attCard)
	clientCode = helperGetClientCodeByName(startingPlayer.Name)

	if err != nil {
		t.Fatalf("Could not get card code from player.Error: %s\n", err.Error())
	} else {
		if err := helperAttack(clientCode, false, attCardCode, http.StatusOK); err != nil {
			t.Fatalf("Could not attack.Error: %s\n", err.Error())
		}
	}

	defendingPlayer := currentGame.GetDefendingPlayer()
	clientCode = helperGetClientCodeByName(defendingPlayer.Name)

	for _, invalidCardCode := range invalidCardCodes {
		// Defending with invalid attacking card code

		if err := helperDefend(clientCode, invalidCardCode, "7D", http.StatusBadRequest); err != nil {
			t.Fatalf("Error: %s\nTest case: Defending with no game started\n", err.Error())
		}
		// Defending with invalid defending card code
		if err := helperDefend(clientCode, attCardCode, invalidCardCode, http.StatusBadRequest); err != nil {
			t.Fatalf("Error: %s\nTest case: Defending with no game started\n", err.Error())
		}
	}


	// Defending with invalid client code
	fakeDefendingCard := &game.Card{Value: 14, Kind: currentGame.KozerCard.Kind}
	currentGame.GetDefendingPlayer().TakeCards(fakeDefendingCard)
	defCardCode, err := game.CardToCode(fakeDefendingCard)
	if err != nil {
		t.Fatalf("Error ocurred when trying to get defending card code. Error: %s\n", err.Error())
	} else {
		if err := helperDefend(invalidClientCode, attCardCode, defCardCode, http.StatusBadRequest); err != nil {
			t.Fatalf("Error: %s\nTest case: Defending with no game started\n", err.Error())
		}
	}

}

func TestTakeCards(t *testing.T) {

	if err := checkMethodsNotAllowed("/takeCards", "POST", takeCards); err != nil {
		t.Fatal(err)
	}

	// Test taking cards when no game created
	validPlayerName := "niv"
	expectedCode := http.StatusBadRequest

	err := helperTakeCards(validPlayerName, expectedCode); if err != nil {
		t.Errorf("Error ocurred when testing for taking cards without game started\n" +
			"Error: %s", err.Error())
	}

	// Create game
	if _, err := helperCreateGame(2, validPlayerName, false, http.StatusOK); err != nil {
		t.Fatalf("could not create game. Error: %s\n", err.Error())
	}

	defer unCreateGame()

	// Start game
	validPlayerName = "niv2"
	if _, err := helperJoinGame(validPlayerName, http.StatusOK); err != nil {
		t.Fatalf("Error ocurred when trying to join (and start) game\n" +
			"Error: %s\n", err.Error())
	}

	// Testing taking cards with a non existing client code
	if err = helperTakeCards(invalidClientCode, expectedCode); err != nil {
		t.Fatalf("could not take cards with un existing player name\n" +
			"Client code used: %s\nError: %s\n", invalidClientCode, err.Error())
	}
}

func TestMoveCardsToBita(t *testing.T) {

	if err := checkMethodsNotAllowed("/moveCardsToBita", "POST", moveCardsToBita); err != nil {
		t.Fatal(err)
	}

	// Test moving cards to bita when no game created
	validPlayerName := "niv"
	expectedCode := http.StatusBadRequest

	err := helperMoveCardsToBita(validPlayerName, expectedCode); if err != nil {
		t.Errorf("Error ocurred when testing for moving cards to bita without game started\n" +
			"Error:: %s", err.Error())
	}

	// Create game
	if _, err := helperCreateGame(2, validPlayerName, false, http.StatusOK); err != nil {
		t.Fatalf("could not create game. Error: %s\n", err.Error())
	}

	defer unCreateGame()

	// Start game
	validPlayerName = "niv2"
	if _, err := helperJoinGame(validPlayerName, http.StatusOK); err != nil {
		t.Fatalf("Error ocurred when trying to join (and start) game\n" +
			"Error: %s\n", err.Error())
	}

	// Testing moving cards to bita with a unknown client code
	if err = helperMoveCardsToBita(invalidClientCode, expectedCode); err != nil {
		t.Fatalf("could not move cards to bita with un existing player name\n" +
			"Client code used: %s\nError: %s\n", invalidClientCode, err.Error())
	}
}

func TestRestartGame(t *testing.T) {

	if err := checkMethodsNotAllowed("/restartGame", "POST", restartGame); err != nil {
		t.Fatal(err)
	}

	validPlayerName := "niv"
	expectedCode := http.StatusBadRequest

	// Test trying to restart when no game is created
	err := helperRestartGame("", expectedCode); if err != nil {
		t.Errorf("Error ocurred when testing for moving cards to bita without game started\n" +
			"Error:: %s", err.Error())
	}

	// Create game
	clientCode, err := helperCreateGame(2, validPlayerName, false, http.StatusOK)
	if err != nil {
		t.Fatalf("could not create game. Error: %s\n", err.Error())
	}

	defer unCreateGame()

	// Testing restarting game when game not started
	if err = helperRestartGame(clientCode, expectedCode); err != nil {
		t.Fatalf("could not restart game when game not started\n" +
			"Client code used: %s\nError: %s\n", clientCode, err.Error())
	}

	// Start game
	validPlayerName = "niv2"
	if _, err := helperJoinGame(validPlayerName, http.StatusOK); err != nil {
		t.Fatalf("Error ocurred when trying to join (and start) game\n" +
			"Error: %s\n", err.Error())
	}

	// Testing restarting game when game not over
	if err = helperRestartGame(clientCode, expectedCode); err != nil {
		t.Fatalf("could not move cards to bita with un existing player name\n" +
			"Client code used: %s\nError: %s\n", clientCode, err.Error())
	}
}

// Internal non-testing methods

func helperCreateInvalidClientCode() string {
	length := configuration.GetInt("ClientIdLength") - 1
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = 'a'
	}
	return string(b)
}

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

func helperCreateGame(playerNum int, name string, shouldUncreate bool, expectedCode int) (string, error) {
	body := httpPayloadTypes.CreateGameRequestObject{
		NumOfPlayers: playerNum,
		PlayerName:   name,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", "/createGame", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("Error occurred: %s\n", err.Error())
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return "", fmt.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	resp := httpPayloadTypes.PlayerJoinedResponse{}
	if expectedCode == http.StatusOK { // Check for proper response
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return "", err
		}
		if resp.PlayerName != name {
			unCreateGame()
			return "", fmt.Errorf("Expected returned name to be %s, instead got %s\n", name, resp.PlayerName)
		}
	} else {
		resp := httpPayloadTypes.ErrorResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return "", err
		}
	}

	if shouldUncreate {
		unCreateGame()
	}

	return resp.IdCode, nil
}

func helperJoinGame(name string, expectedCode int) (string, error) {

	body := httpPayloadTypes.JoinGameRequestObject{
		PlayerName: name,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", "/joinGame", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("Error occurred: %s\n", err.Error())
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(joinGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return "", fmt.Errorf("Join game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	resp := httpPayloadTypes.PlayerJoinedResponse{}
	if expectedCode == http.StatusOK { // Check for proper response
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return "", err
		}
		if resp.PlayerName != name {
			unCreateGame()
			return "", fmt.Errorf("Expected returned name to be %s, instead got %s\n", name, resp.PlayerName)
		}
	} else {
		resp := httpPayloadTypes.ErrorResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return "", err
		}
	}

	return resp.IdCode, nil
}

func helperAttack(clientCode string, shouldUncreateGame bool, cardCode string, expectedCode int) error {
	body := httpPayloadTypes.AttackRequestObject{
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
	req.Header.Set("ConnectionId", clientCode)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(attack)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == http.StatusOK { // Check for proper response
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

func helperLeaveGame(clientCode string, expectedCode int, ) error {
	req, err := http.NewRequest("POST", "/leaveGame", nil)
	if err != nil {
		return err
	}
	req.Header.Set("ConnectionId", clientCode)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(leaveGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Leave game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == http.StatusOK { // Check for proper response
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

func helperDefend(clientCode string, attCardCode string, defCardCode string, expectedCode int) error {

	body := httpPayloadTypes.DefenseRequestObject{
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
	req.Header.Set("ConnectionId", clientCode)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(defend)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == http.StatusOK { // Check for proper response
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

func helperTakeCards(clientCode string, expectedCode int) error {
	req, err := http.NewRequest("POST", "/takeCards", nil)
	if err != nil {
		return fmt.Errorf("Error occurred: %s\n", err.Error())
	}
	req.Header.Set("ConnectionId", clientCode)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(takeCards)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == http.StatusOK { // Check for proper response
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

func helperMoveCardsToBita(clientCode string, expectedCode int) error {
	req, err := http.NewRequest("POST", "/moveCardsToBita", nil)
	if err != nil {
		return fmt.Errorf("Error occurred: %s\n", err.Error())
	}
	req.Header.Set("ConnectionId", clientCode)


	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(moveCardsToBita)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == http.StatusOK { // Check for proper response
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

func helperRestartGame(clientCode string, expectedCode int) error {
	req, err := http.NewRequest("POST", "/restartGame", nil)
	if err != nil {
		return fmt.Errorf("Error occurred: %s\n", err.Error())
	}
	req.Header.Set("ConnectionId", clientCode)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(restartGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		return fmt.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, expectedCode, rr.Body.String())
	}

	jsonResp := rr.Body.Bytes()
	if expectedCode == http.StatusOK { // Check for proper response
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

func helperGetClientCodeByName(name string) string {
	for clientCode, playerName := range clientIdentification {
		if playerName == name {
			return clientCode
		}
	}
	return ""
}