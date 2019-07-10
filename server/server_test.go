package server

import (
	"DurakGo/server/httpPayloadTypes"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGame(t *testing.T) {
	if err := checkMethodsNotAllowed("/createGame", "POST", createGame); err != nil {
		t.Error(err)
	}

	testCases := []struct {
		numOfPlayers int
		name         string
		expectedCode int
	}{
		{numOfPlayers: 0, name: "niv", expectedCode: 400},
		{numOfPlayers: 5, name: "niv", expectedCode: 400},
		{numOfPlayers: 1, name: "niv", expectedCode: 400},
		{numOfPlayers: 3, name: "", expectedCode: 400},
		{numOfPlayers: 3, name: "?", expectedCode: 400},
		{numOfPlayers: 3, name: "|", expectedCode: 400},
		{numOfPlayers: 3, name: "%", expectedCode: 400},
		{numOfPlayers: 3, name: "~", expectedCode: 400},
		{numOfPlayers: 3, name: "/", expectedCode: 400},
		{numOfPlayers: 3, name: "\\", expectedCode: 400},
	}

	for _, testCase := range testCases {
		err := helperCreateGame(testCase.numOfPlayers, testCase.name, true, testCase.expectedCode)
		if err != nil {
			t.Fatalf("Error: %s\nTest Case: %v\n", err.Error(), testCase)
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

	testCases := []struct {
		name         string
		expectedCode int
		create		 bool
	}{
		{name: "niv", expectedCode: 400, create: false},  // Test joining when game not created
		{name: "", expectedCode: 400, create: true},  // Test joining with no name
		{expectedCode: 400, create: true},  // Test joining with name nil
		{name: "?", expectedCode: 400, create: true},  // Illegal name
		{name: "|", expectedCode: 400, create: true},  // Illegal name
		{name: "%", expectedCode: 400, create: true},  // Illegal name
		{name: "~", expectedCode: 400, create: true},  // Illegal name
		{name: "/", expectedCode: 400, create: true},  // Illegal name
		{name: "\\", expectedCode: 400, create: true},  // Illegal name
	}

	for _, testCase := range testCases {

		if testCase.create {
			if err := helperCreateGame(2, "genericCreatorName", false, 200);
			err != nil {
				t.Fatalf("Error: %s\nTest Case: %v\n", err.Error(), testCase)
			}
		}

		if err := helperJoinGame(testCase.name, testCase.expectedCode); err != nil {
			unCreateGame()
			t.Fatalf("Error: %s\nTest Case: %v\n", err.Error(), testCase)
		}

		unCreateGame()
	}

	// Create game with 2 players
	name := "testniv"

	if err := helperCreateGame(2, name, false, 200); err != nil {
		t.Fatalf("Could not create game with 2 players. Error: %s\n", err.Error())
	}

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

	unCreateGame()
}

func TestLeaveGame(t *testing.T) {

	if err := checkMethodsNotAllowed("/leaveGame", "POST", leaveGame); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		leavingPlayerName string
		create            bool
		createNames       []string
		running           bool
		expectedCode	  int
	}{
		{leavingPlayerName: "niv3", create: true, running: true,  // Test leaving when game is running
			createNames: []string{"niv", "niv2"}, expectedCode: 400},
		{leavingPlayerName: "niv", create: false, expectedCode: 400},  // Test leaving when no game created
		{create: true, expectedCode: 400},  // Test leaving with nil
		{leavingPlayerName: "", create: true, expectedCode: 400},  // Test leaving with ""
		{leavingPlayerName: "?", create: true, expectedCode: 400},  // Illegal name
		{leavingPlayerName: "|", create: true, expectedCode: 400},  // Illegal name
		{leavingPlayerName: "%", create: true, expectedCode: 400},  // Illegal name
		{leavingPlayerName: "~", create: true, expectedCode: 400},  // Illegal name
		{leavingPlayerName: "/", create: true, expectedCode: 400},  // Illegal name
		{leavingPlayerName: "\\", create: true, expectedCode: 400},  // Illegal name
		// Testing leaving with a valid name but that does not exist
		{leavingPlayerName: "niv2", create: true, createNames: []string{"niv"}, expectedCode: 400},
	}

	for _, testCase := range testCases {
		if testCase.running {
			if err := helperCreateGameAndJoin(len(testCase.createNames), testCase.createNames, false);
			err != nil {
				t.Errorf("Error: %s\nTest case: %v\n", err.Error(), testCase)
			}
		} else {
			if testCase.create {
				if testCase.createNames != nil {
					if err := helperCreateGameAndJoin(len(testCase.createNames)+1, testCase.createNames, false);
						err != nil {
						t.Errorf("Error: %s\nTest Case: %v\n", err.Error(), testCase)
					}
				} else {
					if err := helperCreateGameAndJoin(2, []string{"genericName"}, false);
						err != nil {
						t.Errorf("Error: %s\nTest case: %v\n", err.Error(), testCase)
					}
				}
			}
		}

		err := helperLeaveGame(testCase.leavingPlayerName, testCase.expectedCode); if err != nil {
			t.Errorf("Error: %s\nTest Case: %v\n", err.Error(), testCase)
		}
		unCreateGame()
	}
}

//func TestAttack(t *testing.T) {
//	if err := checkMethodsNotAllowed("/createGame", "POST", createGame); err != nil {
//		t.Error(err)
//	}
//
//	testCases := []struct {
//		cardCode 	 string
//		name         string
//		expectedCode int
//	}{
//		{}.  // Attacking with no game created
//		{},  // Create game
//		{},  // Attacking with no game started
//		{},  // Start game
//		{},  // Attacking with invalid card code
//		{},  // Attacking with invalid player name
//		{},  // Attacking with non existing player name
//		{},  // Attacking with nil vars
//		{},  // Attack regularly
//	}
//
//	for _, testCase := range testCases {
//		if err := helperAttack(testCase.name, false, testCase.cardCode, testCase.expectedCode); err != nil {
//			t.Fatalf("Error: %s\nTest case: %b\n", err.Error(), testCase)
//		}
//	}
//}

func TestDefend(t *testing.T) {
	// Defending with invalid attacking card code
	// Defending with invalid defending card code
	// Defending with invalid player name
	// Defending with non existing player name
	// Defending when no game created
	// Defending when no game started
	// Defending with nil var
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
		resp := httpPayloadTypes.PlayerJoinedResponse{}
		if err := json.Unmarshal(jsonResp, &resp); err != nil {
			unCreateGame()
			return err
		}
		if resp.PlayerName != name {
			return fmt.Errorf("Expected returned name to be %s, instead got %s\n", name, resp.PlayerName)
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