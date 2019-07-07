package game

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	if _, err := NewGame([]string{"player1", "player2", "player3"}...); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}
}

func TestGetPlayerByName(t *testing.T) {
	playerNames := []string{"player1", "player2", "player3"}
	g, err := NewGame(playerNames...)
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}

	// Valid
	for _, playerName := range playerNames {
		p, err := g.GetPlayerByName(playerName)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}
		if p == nil {
			t.Errorf("Could not get player with name %s\n", playerName)
			return
		}
	}

	// Invalid
	for _, playerName := range []string{"", "PlAyEr1", "%@$#$#", "/////\\\\"} {
		p, err := g.GetPlayerByName(playerName)
		if err == nil || p != nil {
			t.Errorf("Successfully retrieved player with name %s\n", playerName)
			return
		}
	}
}

func TestGetPlayersCardsMap(t *testing.T) {
	g, err := NewGame([]string{"player1", "player2", "player3"}...)
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	g.GetPlayersCardsMap()
}

func TestEndGameWithLoser(t *testing.T) {
	// Handles GetLosingPlayer, IsGameOver and IsDraw methods

	g, err := NewGame([]string{"player1", "player2"}...)
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}

	// Empty deck
	c := g.deck.GetNextCard()
	for c != nil {
		c = g.deck.GetNextCard()
	}

	// Empty player cards
	for _, p := range g.players {
		p.cards = make([]*Card, 0)
	}

	// Set up player cards
	p1, err := g.GetPlayerByName("player1")
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	c1 := makeCard("Clubs", 13)
	p1.TakeCards(c1)

	p2, err := g.GetPlayerByName("player1")
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	c2 := makeCard("Clubs", 9)
	p2.TakeCards(c2)

	// Set up turn
	g.startingPlayer = p1
	g.defendingPlayer = p2

	if err = g.Attack(p1, c1); err != nil {
		t.Errorf("Error occurred while attacking: %s\n", err.Error())
		return
	}

	// Test invalid
	p := g.GetLosingPlayer()
	if p != nil {
		t.Errorf("Receieved losing player %s\n", p.Name)
		return
	}
	if g.IsGameOver() {
		t.Errorf("Game should not be over\n")
		return
	}

	if g.IsDraw() {
		t.Errorf("Game should not be over\n")
		return
	}

	if err := g.PickUpCards(); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}

	// Test valid
	if !g.IsGameOver() {
		t.Errorf("Game should be over\n")
		return
	}
	if g.IsDraw() {
		t.Errorf("Game should not be draw\n")
		return
	}
	p = g.GetLosingPlayer()
	if p == nil {
		t.Errorf("Did not receieve any losing player\n")
		return
	}
}

func TestEndGameWithDraw(t *testing.T) {
	// Handles GetLosingPlayer, IsGameOver and IsDraw methods

	g, err := NewGame([]string{"player1", "player2"}...)
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}

	// Empty deck
	c := g.deck.GetNextCard()
	for c != nil {
		c = g.deck.GetNextCard()
	}

	// Empty player cards
	for _, p := range g.players {
		p.cards = make([]*Card, 0)
	}

	// Set up player cards
	p1, err := g.GetPlayerByName("player1")
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	c1 := makeCard("Clubs", 13)
	p1.TakeCards(c1)

	p2, err := g.GetPlayerByName("player1")
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	c2 := makeCard("Clubs", 9)
	p2.TakeCards(c2)

	// Set up turn
	g.startingPlayer = p2
	g.defendingPlayer = p1

	if err = g.Attack(p2, c2); err != nil {
		t.Errorf("Error occurred while attacking: %s\n", err.Error())
		return
	}

	// Test invalid
	p := g.GetLosingPlayer()
	if p != nil {
		t.Errorf("Receieved losing player %s\n", p.Name)
		return
	}
	if g.IsGameOver() {
		t.Errorf("Game should not be over\n")
		return
	}

	if g.IsDraw() {
		t.Errorf("Game should not be over\n")
		return
	}

	if err = g.Defend(p1, c2, c1); err != nil {
		t.Errorf("Error ocurred while trying to defend: %s\n", err.Error())
		return
	}

	if err = g.MoveToBita(); err != nil {
		t.Errorf("Error occurred while moving cards to bita: %s\n", err.Error())
		return
	}

	// Test valid
	if !g.IsGameOver() {
		t.Errorf("Game should be over\n")
		return
	}
	if !g.IsDraw() {
		t.Errorf("Game should be draw\n")
		return
	}
	p = g.GetLosingPlayer()
	if p != nil {
		t.Errorf("Received a losing player %s while draw is expectedr\n", p.Name)
		return
	}

}

func TestPickUpCards(t *testing.T) {
	playerNames := []string{"player1", "player2"}
	g, err := NewGame(playerNames...)
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}

	// test with empty board
	b := g.board
	if err := g.PickUpCards(); err == nil {
		t.Errorf("Expected error for an empty board\n")
		return
	}

	// test with one undefended card
	defendingPlayer := g.GetDefendingPlayer()
	b = NewBoard()
	c1 := makeCard("Clubs", 9)
	b.AddAttackingCard(c1)
	g.board = b

	if err := g.PickUpCards(); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	success := false
	for _, c := range defendingPlayer.PeekCards() {
		if c == c1 {
			success = true
		}
	}
	if !success {
		t.Errorf("could not find %v in player's hand\nReceieved hand:%v\n", c1, defendingPlayer.PeekCards())
		return
	}

	// test with one defended card

	defendingPlayer = g.GetDefendingPlayer()
	b = NewBoard()
	c2 := makeCard("Clubs", 10)
	b.AddAttackingCard(c1)
	if err := b.AddDefendingCard(c1, c2); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	g.board = b

	if err := g.PickUpCards(); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	success1, success2 := false, false
	for _, c := range defendingPlayer.PeekCards() {
		if c == c1 {
			success1 = true
		}
		if c == c2 {
			success2 = true
		}
	}
	if !(success2 && success1) {
		t.Errorf("could not find cards %v and %v in player's hand\nReceieved hand: %v\n", c1, c2, defendingPlayer.PeekCards())
	}

	// test with one defended and one undefended

	defendingPlayer = g.GetDefendingPlayer()
	b = NewBoard()
	c3 := makeCard("Hearts", 13)
	b.AddAttackingCard(c1)
	if err := b.AddDefendingCard(c1, c2); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	b.AddAttackingCard(c3)
	g.board = b

	if err := g.PickUpCards(); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	success1, success2, success3 := false, false, false
	for _, c := range defendingPlayer.PeekCards() {
		if c == c1 {
			success1 = true
		}
		if c == c2 {
			success2 = true
		}
		if c == c3 {
			success3 = true
		}
	}
	if !(success2 && success1 && success3) {
		t.Errorf("could not find cards %v, %v and %v in player's hand\nReceieved hand: %v\n", c1, c2, c3, defendingPlayer.PeekCards())
	}

}

func TestMoveToBita(t *testing.T) {

	playerNames := []string{"player1", "player2"}
	g, err := NewGame(playerNames...)
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}

	// test with empty board
	b := g.board
	if err := g.MoveToBita(); err == nil {
		t.Errorf("Expected error for an empty board\n")
		return
	}

	// test with one undefended card
	b = NewBoard()
	c1 := makeCard("Clubs", 9)
	b.AddAttackingCard(c1)
	g.board = b

	if err := g.MoveToBita(); err == nil {
		t.Errorf("Expected error\n")
		return
	}
	// test with one defended card

	b = NewBoard()
	c2 := makeCard("Clubs", 10)
	b.AddAttackingCard(c1)
	if err := b.AddDefendingCard(c1, c2); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	g.board = b

	if err := g.MoveToBita(); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	// test with one defended and one undefended

	b = NewBoard()
	c3 := makeCard("Hearts", 13)
	b.AddAttackingCard(c1)
	if err := b.AddDefendingCard(c1, c2); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
		return
	}
	b.AddAttackingCard(c3)
	g.board = b

	if err := g.MoveToBita(); err == nil {
		t.Errorf("Expected error\n")
		return
	}

}

func TestDefend(t *testing.T) {

	kozerCard := makeCard("Hearts", 9)
	c1 := makeCard("Clubs", 10)
	c2 := makeCard("Diamonds", 10)
	c3 := makeCard("Clubs", 9)
	c4 := makeCard("Hearts", 14)
	c5 := makeCard("Hearts", 6)

	invalidTestCases := []struct {
		att		*Card
		def		*Card
		b		*Board
		pCards	[]*Card
	}{
		{att: c3, b: NewBoard(), def: c1, pCards: []*Card{c1}},  // Test empty board
		{att: GetRandomCard(), def: c4, b: NewBoard(), pCards: []*Card{c4}},  // Test attacking card not on board
		{att: c1, def: c5, b: getBoardWithAttackingCardsOnly(c1), pCards: []*Card{c2,c3,c4}},  // Test defending card not in hand
		{att: c1, def: c4, b: getBoardWithDefendedCard(c1, GetRandomCard()), pCards: []*Card{c4}}, // Test attacking card already responded
		{att: c1, def: c2, b: getBoardWithAttackingCardsOnly(c1), pCards: []*Card{c2}},  // Test card can not answer wrong suit
		{att: c4, def: c5, b: getBoardWithAttackingCardsOnly(c4), pCards: []*Card{c5}},  // Test card can not answer lower kozer
		{att: c1, def: c3, b: getBoardWithAttackingCardsOnly(c1), pCards: []*Card{c3}},  // Test card can not answer lower than attacking
		{att: c5, def: c2, b: getBoardWithAttackingCardsOnly(c5), pCards: []*Card{c2}},  // Test card can not answer att kozer def non kozer
	}

	validTestCases := []struct {
		att		*Card
		b      *Board
		pCards []*Card
	}{
		{att: c3, b: getBoardWithAttackingCardsOnly(c3), pCards: []*Card{c1}},  // Test valid non kozer
		{att: c5, b: getBoardWithAttackingCardsOnly(c5), pCards: []*Card{c4}},  // Test valid both kozers
		{att: c2, b: getBoardWithAttackingCardsOnly(c2), pCards: []*Card{c5}},  // Test valid defending kozer
	}

	for _, testCase := range invalidTestCases {
		playerNames := []string{"player1", "player2"}
		g, err := NewGame(playerNames...)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}
		g.KozerCard = kozerCard

		g.board = testCase.b
		p := g.GetDefendingPlayer()
		p.cards = testCase.pCards
		defendingCard := testCase.def

		attackingCard := testCase.att
		if err := g.Defend(p, attackingCard, defendingCard); err == nil {
			t.Errorf("Expected error\nBoard: %v\nPlayer's cards: %v\n", testCase.b, p.cards)
		}
	}

	for _, testCase := range validTestCases {
		playerNames := []string{"player1", "player2"}
		g, err := NewGame(playerNames...)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}
		g.KozerCard = kozerCard

		g.board = testCase.b
		p := g.GetDefendingPlayer()
		p.cards = testCase.pCards
		if err := g.Defend(p, testCase.att, p.cards[0]); err != nil {
			t.Errorf("Error occurred: %s\nBoard: %v\nPlayer's cards: %v\n", err.Error(), testCase.b, p.cards)
		}
	}
}

func TestAttack(t *testing.T) {

	c1 := makeCard("Hearts", 6)
	c2 := makeCard("Hearts", 8)
	c3 := makeCard("Diamonds", 8)

	invalidTestCases := []struct {
		playerNames				[]string
		startingPlayerName		string
		defendingPlayerName		string
		board					*Board
		attackingPlayerName		string
		attackingPlayerCards	[]*Card
		defendingPlayerCards	[]*Card
		attackCard				*Card
	}{
		// empty board, correct player turn, card is nil, player hand empty
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: NewBoard(), attackingPlayerName: "player1"},

		// empty board, correct player turn, card is nil, player hand not empty
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: NewBoard(), attackingPlayerName: "player1", attackingPlayerCards: []*Card{c1}},

		// empty board, wrong player turn
		{playerNames: []string{"player1", "player2", "player3"}, startingPlayerName: "player1",
			defendingPlayerName: "player2", board: NewBoard(), attackingPlayerName: "player3",
			attackingPlayerCards: []*Card{c1}, attackCard: c1},

		// not empty board, card not valid
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: getBoardWithAttackingCardsOnly(c1), attackingPlayerName: "player1", attackingPlayerCards: []*Card{c2},
			attackCard: c2},

		// not empty board, defending player trying to add
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: getBoardWithAttackingCardsOnly(c2), attackingPlayerName: "player2", attackingPlayerCards: []*Card{c3},
			attackCard: c3},

		// empty board, defending player trying to add
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1",
			defendingPlayerName: "player2", board: NewBoard(), attackingPlayerName: "player2",
			attackingPlayerCards: []*Card{c1}, attackCard: c1},

		// non empty board, defending player limit reached on table
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: getBoardWithAttackingCardsOnly(GetRandomCard()), attackingPlayerName: "player1",
				attackingPlayerCards: []*Card{c1}, defendingPlayerCards: []*Card{GetRandomCard()}, attackCard: c1},

		// non empty board, global card limit reached
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: getBoardWithAttackingCardsOnly(GetRandomCard(), GetRandomCard(), GetRandomCard(), GetRandomCard(),
				GetRandomCard(), GetRandomCard()), attackingPlayerName: "player1",
			attackingPlayerCards: []*Card{c1}, attackCard: c1},

		// empty board, correct player, card not in hand
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: NewBoard(), attackingPlayerName: "player1",
			attackingPlayerCards: []*Card{c2}, attackCard: c1},

		// non empty board, card not in hand
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: getBoardWithAttackingCardsOnly(GetRandomCard()), attackingPlayerName: "player1",
			attackingPlayerCards: []*Card{c2}, attackCard: c1},

		// non empty board, cant add since number is not on board
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: getBoardWithAttackingCardsOnly(c1), attackingPlayerName: "player1",
			attackingPlayerCards: []*Card{c2}, attackCard: c2},
	}

	validTestCases := []struct {
		playerNames				[]string
		startingPlayerName		string
		defendingPlayerName		string
		board					*Board
		attackingPlayerName		string
		attackingPlayerCards	[]*Card
		defendingPlayerCards	[]*Card
		attackCard				*Card
	}{
		// test valid empty board + player starting
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: NewBoard(), attackingPlayerName: "player1", attackingPlayerCards: []*Card{c1}, attackCard: c1},

		// non empty board, non starting player adds
		{playerNames: []string{"player1", "player2", "player3"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: getBoardWithAttackingCardsOnly(c2), attackingPlayerName: "player3",
			attackingPlayerCards: []*Card{c3}, attackCard: c3},

		// non empty board, starting player adds
		{playerNames: []string{"player1", "player2"}, startingPlayerName: "player1", defendingPlayerName: "player2",
			board: getBoardWithAttackingCardsOnly(c2), attackingPlayerName: "player1",
			attackingPlayerCards: []*Card{c3}, attackCard: c3},
	}

	for _, testCase := range invalidTestCases {
		g, err := NewGame(testCase.playerNames...)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}

		g.board = testCase.board

		startingPlayer, err := g.GetPlayerByName(testCase.startingPlayerName)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}
		g.startingPlayer = startingPlayer

		defendingPlayer, err := g.GetPlayerByName(testCase.defendingPlayerName)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}
		g.defendingPlayer = defendingPlayer

		if testCase.defendingPlayerCards != nil {
			defendingPlayer.cards = testCase.defendingPlayerCards
		}

		attackingPlayer, err := g.GetPlayerByName(testCase.attackingPlayerName)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}
		attackingPlayer.cards = testCase.attackingPlayerCards

		attackingCard := testCase.attackCard
		if err := g.Attack(attackingPlayer, attackingCard); err == nil {
			t.Errorf("Expected error\nBoard: %v\nStarting: %s\nDefending: %s\n%s attacking with %v\n",
				testCase.board, startingPlayer, defendingPlayer, attackingPlayer, attackingCard)
		}
	}

	for _, testCase := range validTestCases {
		g, err := NewGame(testCase.playerNames...)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}

		g.board = testCase.board

		startingPlayer, err := g.GetPlayerByName(testCase.startingPlayerName)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}
		g.startingPlayer = startingPlayer

		defendingPlayer, err := g.GetPlayerByName(testCase.defendingPlayerName)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}
		g.defendingPlayer = defendingPlayer

		if testCase.defendingPlayerCards != nil {
			defendingPlayer.cards = testCase.defendingPlayerCards
		}

		attackingPlayer, err := g.GetPlayerByName(testCase.attackingPlayerName)
		if err != nil {
			t.Errorf("Error occurred: %s\n", err.Error())
			return
		}
		attackingPlayer.cards = testCase.attackingPlayerCards

		attackingCard := testCase.attackCard
		if err := g.Attack(attackingPlayer, attackingCard); err != nil {
			t.Errorf("Error occurred: %s\n\nBoard: %v\nStarting: %s\nDefending: %s\n%s attacking with %v\n",
				err.Error(), testCase.board, startingPlayer, defendingPlayer, attackingPlayer, attackingCard)
		}
	}


}