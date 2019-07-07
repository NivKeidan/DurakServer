package game

import (
	"DurakGo/game"
	"testing"
)

func TestIsEmpty(t *testing.T) {
	b := game.NewBoard()
	if !b.IsEmpty() {
		t.Errorf("Board is not empty")
	}

	c := GetRandomCard()
	b.AddAttackingCard(c)
	if b.IsEmpty() {
		t.Errorf("Board should be empty")
	}
}

func TestEmptyBoard(t *testing.T) {
	b := game.NewBoard()
	c := GetRandomCard()
	b.AddAttackingCard(c)
	if b.IsEmpty() {
		t.Errorf("Board should not be empty")
	}
	b.EmptyBoard()
	if !b.IsEmpty() {
		t.Errorf("Board should be empty")
	}
}

func TestAddAttackingCard(t *testing.T) {
	b := game.NewBoard()
	c1 := GetRandomCard()
	b.AddAttackingCard(c1)
	c2 := GetRandomCard()
	b.AddAttackingCard(c2)
	c3 := GetRandomCard()
	b.AddAttackingCard(c3)
	c4 := GetRandomCard()
	b.AddAttackingCard(c4)
	expectedCount := 4
	counter := 0

	for _, card := range b.PeekCards() {
		counter++
		if card != c1 && card != c2 && card != c3 && card != c4 {
			t.Errorf("Got unknown card %v on board\n", card)
		}
	}

	if counter != expectedCount {
		t.Errorf("Counter reached %d instead of %d\nCards returned: %v\n", counter, expectedCount, b.PeekCards())
	}

}

func TestDefendCard(t *testing.T) {
	b := game.NewBoard()
	kozerKind := game.Kind("Hearts")
	cardPairsToTest := make([]cardPair, 0)

	// Valid pairs
	cardPairsToTest = []cardPair{
		{makeCard("Clubs", 13), makeCard("Clubs", 14)},
		{makeCard("Clubs", 2), makeCard("Clubs", 13)},
		{makeCard("Hearts", 13), makeCard("Hearts", 14)},
		{makeCard("Clubs", 14), makeCard("Hearts", 2)},
		{makeCard("Clubs", 2), makeCard("Hearts", 2)},
		{makeCard("Clubs", 2), makeCard("Hearts", 10)},
	}

	for _, cardPair := range cardPairsToTest {
		b.AddAttackingCard(cardPair.attackingCard)
		if err := b.DefendCard(cardPair.attackingCard, cardPair.defendingCard, &kozerKind); err != nil {
			t.Errorf("Could not defend %v with %v\n", cardPair.attackingCard, cardPair.defendingCard)
		}
		b.EmptyBoard()
	}

	// Invalid pairs
	cardPairsToTest = []cardPair{
		{makeCard("Clubs", 14), makeCard("Clubs", 13)},
		{makeCard("Clubs", 2), makeCard("Clubs", 2)},
		{makeCard("Hearts", 14), makeCard("Hearts", 13)},
		{makeCard("Hearts", 2), makeCard("Clubs", 2)},
		{makeCard("Hearts", 2), makeCard("Clubs", 3)},
	}
	for _, cardPair := range cardPairsToTest {
		b.AddAttackingCard(cardPair.attackingCard)
		if err := b.DefendCard(cardPair.attackingCard, cardPair.defendingCard, &kozerKind); err == nil {
			t.Errorf("Defended %v with %v successfully\n", cardPair.attackingCard, cardPair.defendingCard)
		}
		b.EmptyBoard()
	}

}

func TestCanCardBeAdded(t *testing.T) {
	b := game.NewBoard()
	b.AddAttackingCard(makeCard("Clubs", 2))
	b.AddAttackingCard(makeCard("Diamonds", 2))
	b.AddAttackingCard(makeCard("Hearts", 9))
	b.AddAttackingCard(makeCard("Clubs", 10))
	b.AddAttackingCard(makeCard("Clubs", 13))
	b.AddAttackingCard(makeCard("Clubs", 14))
	b.AddAttackingCard(makeCard("Clubs", 11))

	// Valid options
	for _, i := range []int{2,9,10,11,13,14} {
		c := makeCard("Clubs", i)
		if !b.CanCardBeAdded(c) {
			t.Errorf("Receieved false for being able to add %v to board with: %v\n", c, b.PeekCards())
		}
	}

	// Invalid options
	for _, i := range []int{0, 1, 3,4,5,6,7,8,12, 15, 16, 100, -24, -0, 150} {
		c := makeCard("Clubs", i)
		if b.CanCardBeAdded(c) {
			t.Errorf("Receieved true for being able to add %v to board with: %v\n", c, b.PeekCards())
		}
	}
}

func TestIsCardLimitReached(t *testing.T) {
	// TODO Integrate test options object to test max cards with
}

func TestAreAllCardsDefended(t *testing.T) {
	b := game.NewBoard()
	kozerKind := game.Kind("Hearts")

	att := makeCard("Clubs", 3)
	def := makeCard("Clubs", 5)
	b.AddAttackingCard(att)
	if err := b.DefendCard(att, def, &kozerKind); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}

	if !b.AreAllCardsDefended() {
		t.Errorf("Should be true. Board: %v\n", b)
	}

	att = makeCard("Clubs", 10)
	def = makeCard("Hearts", 2)
	b.AddAttackingCard(att)
	if err := b.DefendCard(att, def, &kozerKind); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}

	if !b.AreAllCardsDefended() {
		t.Errorf("Should be true. Board: %v\n", b)
	}

	att = makeCard("Spades", 12)
	b.AddAttackingCard(att)

	if b.AreAllCardsDefended() {
		t.Errorf("Should be false. Board: %v\n", b)
	}
}

func TestGetAllCards(t *testing.T) {
	b := game.NewBoard()
	kozerKind := game.Kind("Hearts")
	att := makeCard("Clubs", 3)
	def := makeCard("Clubs", 5)
	att2 := makeCard("Clubs", 10)
	def2 := makeCard("Hearts", 2)
	att3 := makeCard("Spades", 12)
	expectedCount := 5
	counter := 0

	b.AddAttackingCard(att)
	if err := b.DefendCard(att, def, &kozerKind); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}

	b.AddAttackingCard(att2)
	if err := b.DefendCard(att2, def2, &kozerKind); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}

	b.AddAttackingCard(att3)

	for _, card := range b.PeekCards() {
		counter++
		if card != att && card != att2 && card != att3 && card != def && card != def2 {
			t.Errorf("Got unknown card %v on board\n", card)
		}
	}

	if counter != expectedCount {
		t.Errorf("Counter reached %d instead of %d\n", counter, expectedCount)
	}

}

func TestGetAllCardsOnBoard(t *testing.T) {
	b := game.NewBoard()
	kozerKind := game.Kind("Hearts")
	att := makeCard("Clubs", 3)
	def := makeCard("Clubs", 5)
	att2 := makeCard("Clubs", 10)
	def2 := makeCard("Hearts", 2)
	att3 := makeCard("Spades", 12)
	expectedCount := 3
	counter := 0

	b.AddAttackingCard(att)
	if err := b.DefendCard(att, def, &kozerKind); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}

	b.AddAttackingCard(att2)
	if err := b.DefendCard(att2, def2, &kozerKind); err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}

	b.AddAttackingCard(att3)

	for _, cardOnBoard := range b.PeekCardsOnBoard() {
		counter++
		a := cardOnBoard.GetAttackingCard()
		d := cardOnBoard.GetDefendingCard()
		if !((a == att && d == def) || (a == att2 && d == def2) || (a == att3 && d == nil)) {
			t.Errorf("Got unknown cardOnBoard: %v\n", cardOnBoard)
		}
	}

	if counter != expectedCount {
		t.Errorf("Counter reached %d instead of %d\n", counter, expectedCount)
	}
}