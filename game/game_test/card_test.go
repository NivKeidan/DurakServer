package game

import (
	"DurakGo/game"
	"fmt"
	"strconv"
	"testing"
)

func TestNewCard(t *testing.T) {
	// Valid pairs
	cardDataToTest := []*cardData{makeCardData("Clubs", 2), makeCardData("Diamonds", 13)}
	for _, cardData := range cardDataToTest {
		_, err := game.NewCard(cardData.k, cardData.v)
		if err != nil {
			t.Errorf("Card value %d of kind %s not valid\nError: %s\n", cardData.v, cardData.k, err.Error())
		}
	}

	// Invalid pairs

	cardDataToTest = []*cardData{makeCardData("Clubs", 0), makeCardData("Clubs", 15),
		makeCardData("Clubs", -51), makeCardData("clubs", 1), makeCardData("omg", 13)}
	for _, cardData := range cardDataToTest {
		_, err := game.NewCard(cardData.k, cardData.v)
		if err == nil {
			t.Errorf("Card value %d of kind %s not valid\n", cardData.v, cardData.k)
		}
	}
}

func TestNewCardByCode(t *testing.T) {
	// Valid cases

	testCodes := []string{"10D", "8S", "KH", "2C", "AC", "JD"}
	for _, code := range testCodes {
		_, err := game.NewCardByCode(code)
		if err != nil {
			t.Errorf("code %s was invalid\nerror: %s\n", code, err.Error())
		}
	}

	// Invalid cases

	testCodes = []string{"01D", "014C", "1c", "", "1", "q", "c", "-15", "?", "%"}
	for _, code := range testCodes {
		_, err := game.NewCardByCode(code)
		if err == nil {
			t.Errorf("code %s was valid!\n", code)
		}
	}
}

func TestCardToCode(t *testing.T) {

	for n := 2; n < 15; n++ {
		for _, c := range []string{"Clubs", "Hearts", "Diamonds", "Spades"} {
			var valueString string
			if n <= 10 {
				valueString = strconv.Itoa(n)
			} else {
				switch n {
				case 11:
					valueString = "J"
				case 12:
					valueString = "Q"
				case 13:
					valueString = "K"
				case 14:
					valueString = "A"
				}
			}
			codeExpected := fmt.Sprintf("%s%s", valueString, string(c[0]))
			card := makeCard(c, n)
			codeReceived, err := game.CardToCode(card)
			if err != nil {
				t.Errorf("Error occurred: %s\n", err.Error())
			}
			if codeExpected != codeReceived {
				t.Errorf("Expected code: %s, Receieved code: %s\n", codeExpected, codeReceived)
			}
		}
	}
}

func TestCanDefendCard(t *testing.T) {
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
		if !cardPair.defendingCard.CanDefendCard(cardPair.attackingCard, &kozerKind) {
			t.Errorf("%v could not defend %v\n", cardPair.defendingCard, cardPair.attackingCard)
		}
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
		if cardPair.defendingCard.CanDefendCard(cardPair.attackingCard, &kozerKind) {
			t.Errorf("%v can defend %v\n", cardPair.defendingCard, cardPair.attackingCard)
		}
	}
}