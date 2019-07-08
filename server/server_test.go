package server

import (
	"DurakGo/server/httpPayloadObjects"
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
		body := httpPayloadObjects.CreateGameRequestObject{
			NumOfPlayers: testCase.numOfPlayers,
			PlayerName:   testCase.name,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
		}
		req, err := http.NewRequest("POST", "/createGame", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Error occurred: %s\n", err.Error())
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(createGame)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != testCase.expectedCode {
			t.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
				status, testCase.expectedCode, rr.Body.String())
		}
	}

	// Test creating more than one game
	body1 := httpPayloadObjects.CreateGameRequestObject{
		NumOfPlayers: 3,
		PlayerName:   "player1",
	}

	jsonBody1, err := json.Marshal(body1)
	if err != nil {
		t.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", "/createGame", bytes.NewBuffer(jsonBody1))
	if err != nil {
		t.Fatalf("Error occurred: %s\n", err.Error())
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != 200 {
		t.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, 200, rr.Body.String())
	}

	body2 := httpPayloadObjects.CreateGameRequestObject{
		NumOfPlayers: 3,
		PlayerName:   "player2",
	}

	jsonBody2, err := json.Marshal(body2)
	if err != nil {
		t.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req2, err := http.NewRequest("POST", "/createGame", bytes.NewBuffer(jsonBody2))
	if err != nil {
		t.Fatalf("Error occurred: %s\n", err.Error())
	}

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if status := rr2.Code; status != 400 {
		t.Errorf("Create game handler returned wrong status code: got %v want %v\nResponse:%s\n",
			status, 400, rr2.Body.String())
	}
	unCreateGame()
}

func TestCreateGame2Players( t *testing.T) {
	err := testCreateGameXPlayers(2, "niv", true); if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestCreateGame3Players( t *testing.T) {
	err := testCreateGameXPlayers(3, "niv", true); if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestCreateGame4Players( t *testing.T) {
	err := testCreateGameXPlayers(4, "niv", true); if err != nil {
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
	}{
		{name: "niv", expectedCode: 400},  // Test joining when game not created
		{name: "", expectedCode: 400},  // Test joining with no name
		{expectedCode: 400},  // Test joining with name nil
		{name: "?", expectedCode: 400},  // Illegal name
		{name: "|", expectedCode: 400},  // Illegal name
		{name: "%", expectedCode: 400},  // Illegal name
		{name: "~", expectedCode: 400},  // Illegal name
		{name: "/", expectedCode: 400},  // Illegal name
		{name: "\\", expectedCode: 400},  // Illegal name
	}

	for _, testCase := range testCases {

		body := httpPayloadObjects.JoinGameRequestObject{
			PlayerName: testCase.name,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
		}
		req, err := http.NewRequest("POST", "/joinGame", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Error occurred: %s\n", err.Error())
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(joinGame)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != testCase.expectedCode {
			t.Errorf("Join game handler returned wrong status code: got %v want %v\nResponse: %s\n",
				status, testCase.expectedCode, rr.Body.String())
		}
	}

	// Create game with 2 players
	name := "testniv"
	expectedCode := 200

	if err := testCreateGameXPlayers(2, name, false); err != nil {
		t.Fatalf(err.Error())
	}

	// Test join player with same name used for creation
	body := httpPayloadObjects.JoinGameRequestObject{
		PlayerName: name,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", "/joinGame", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Error occurred: %s\n", err.Error())
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(joinGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != 400 {
		t.Errorf("Join game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, 400, rr.Body.String())
	}

	// Join second player properly
	body = httpPayloadObjects.JoinGameRequestObject{
		PlayerName: "testniv3",
	}
	jsonBody, err = json.Marshal(body)
	if err != nil {
		t.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err = http.NewRequest("POST", "/joinGame", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Error occurred: %s\n", err.Error())
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(joinGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedCode {
		t.Errorf("Join game handler returned wrong status code: got %v want %v\n Response: %s\n",
			status, expectedCode, rr.Body.String())
	}

	// Test try joining a running game
	body = httpPayloadObjects.JoinGameRequestObject{
		PlayerName: name,
	}
	jsonBody, err = json.Marshal(body)
	if err != nil {
		t.Errorf("Could not JSONify request object. Error: %s\n", err.Error())
	}
	req, err = http.NewRequest("POST", "/joinGame", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Error occurred: %s\n", err.Error())
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(joinGame)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != 400 {
		t.Errorf("Join game handler returned wrong status code: got %v want %v\nResponse: %s\n",
			status, 400, rr.Body.String())
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
			if err := testCreateGameAndJoin(len(testCase.createNames), testCase.createNames, false);
			err != nil {
				t.Errorf("Error: %s\nTest case: %v\n", err.Error(), testCase)
			}
		} else {
			if testCase.create {
				if testCase.createNames != nil {
					if err := testCreateGameAndJoin(len(testCase.createNames)+1, testCase.createNames, false);
						err != nil {
						t.Errorf("Error: %s\nTest Case: %v\n", err.Error(), testCase)
					}
				} else {
					if err := testCreateGameAndJoin(2, []string{"genericName"}, false);
						err != nil {
						t.Errorf("Error: %s\nTest case: %v\n", err.Error(), testCase)
					}
				}
			}
		}

		body := httpPayloadObjects.LeaveGameRequestObject{
			PlayerName: testCase.leavingPlayerName,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Errorf("Could not JSONify request object. Error: %s\nTest Case: %v\n", err.Error(), testCase)
		}
		req, err := http.NewRequest("POST", "/leaveGame", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Error occurred: %s\nTest case: %v\n", err.Error(), testCase)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(leaveGame)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != testCase.expectedCode {
			t.Errorf("Leave game handler returned wrong status code: got %v want %v\nResponse: %s\nTest case: %v\n",
				status, testCase.expectedCode, rr.Body.String(), testCase)
		}

		unCreateGame()
	}
}

func TestAttack(t *testing.T) {

}

func TestDefend(t *testing.T) {

}

func TestTakeCards(t *testing.T) {

}

func TestMoveCardsToBita(t *testing.T) {

}

func TestRestartGame(t *testing.T) {

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

func testCreateGameXPlayers(n int, name string, shouldUncreate bool) error {
	expectedCode := 200
	body := httpPayloadObjects.CreateGameRequestObject{
		NumOfPlayers: n,
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
	resp := httpPayloadObjects.PlayerJoinedResponse{}
	if err := json.Unmarshal(jsonResp, &resp); err != nil {
		unCreateGame()
		return err
	}
	if resp.PlayerName != name {
		unCreateGame()
		return fmt.Errorf("Expected returned name to be %s, instead got %s\n", name, resp.PlayerName)
	}

	if shouldUncreate {
		unCreateGame()
	}

	return nil
}

func testJoinGame(name string) error {
	expectedCode := 200

	body := httpPayloadObjects.JoinGameRequestObject{
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

	return nil
}

func testCreateGameAndJoin(playerNum int, playerNames []string, shouldUncreate bool) error {
	// Create game
	err := testCreateGameXPlayers(playerNum, playerNames[0], false); if err != nil {
		return err
	}

	for i := 1; i < len(playerNames); i++ {
		if err := testJoinGame(playerNames[i]); err != nil {
			unCreateGame()
			return err
		}
	}

	if shouldUncreate {
		unCreateGame()
	}

	return nil

}
