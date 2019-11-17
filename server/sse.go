package server

import "DurakGo/server/httpPayloadTypes"

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
