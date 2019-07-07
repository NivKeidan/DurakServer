package game

import (
	"fmt"
	"strconv"
	"testing"
)

func TestNewCard(t *testing.T) {
	// Valid pairs
	cardDataToTest := []*cardData{makeCardData("Clubs", 6), makeCardData("Diamonds", 13)}
	for _, cardData := range cardDataToTest {
		_, err := NewCard(cardData.k, cardData.v)
		if err != nil {
			t.Errorf("Card value %d of kind %s not valid\nError: %s\n", cardData.v, cardData.k, err.Error())
		}
	}

	// Invalid pairs

	cardDataToTest = []*cardData{makeCardData("Clubs", 0), makeCardData("Clubs", 15),
		makeCardData("Clubs", -51), makeCardData("clubs", 1), makeCardData("omg", 13)}
	for _, cardData := range cardDataToTest {
		_, err := NewCard(cardData.k, cardData.v)
		if err == nil {
			t.Errorf("Card value %d of kind %s not valid\n", cardData.v, cardData.k)
		}
	}
}

func TestNewCardByCode(t *testing.T) {
	// Valid cases

	testCodes := []string{"10D", "8S", "KH", "6C", "AC", "JD"}
	for _, code := range testCodes {
		_, err := NewCardByCode(code)
		if err != nil {
			t.Errorf("code %s was invalid\nerror: %s\n", code, err.Error())
		}
	}

	// Invalid cases

	testCodes = []string{"1C", "2D", "3S", "4D", "5H", "01D", "014C", "1c", "", "1", "q", "c", "-15", "?", "%"}
	for _, code := range testCodes {
		_, err := NewCardByCode(code)
		if err == nil {
			t.Errorf("code %s was valid!\n", code)
		}
	}
}

func TestCardToCode(t *testing.T) {

	for n := 6; n < 15; n++ {
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
			codeReceived, err := CardToCode(card)
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
	kozerKind := Kind("Hearts")
	cardPairsToTest := make([]cardPair, 0)

	// Valid pairs
	cardPairsToTest = []cardPair{
		{makeCard("Clubs", 13), makeCard("Clubs", 14)},
		{makeCard("Clubs", 6), makeCard("Clubs", 13)},
		{makeCard("Hearts", 13), makeCard("Hearts", 14)},
		{makeCard("Clubs", 14), makeCard("Hearts", 6)},
		{makeCard("Clubs", 6), makeCard("Hearts", 6)},
		{makeCard("Clubs", 6), makeCard("Hearts", 10)},
	}

	for _, cardPair := range cardPairsToTest {
		if !cardPair.defendingCard.CanDefendCard(cardPair.attackingCard, &kozerKind) {
			t.Errorf("%v could not defend %v\n", cardPair.defendingCard, cardPair.attackingCard)
		}
	}

	// Invalid pairs
	cardPairsToTest = []cardPair{
		{makeCard("Clubs", 14), makeCard("Clubs", 13)},
		{makeCard("Clubs", 6), makeCard("Clubs", 6)},
		{makeCard("Hearts", 14), makeCard("Hearts", 13)},
		{makeCard("Hearts", 6), makeCard("Clubs", 6)},
		{makeCard("Hearts", 6), makeCard("Clubs", 7)},
	}
	for _, cardPair := range cardPairsToTest {
		if cardPair.defendingCard.CanDefendCard(cardPair.attackingCard, &kozerKind) {
			t.Errorf("%v can defend %v\n", cardPair.defendingCard, cardPair.attackingCard)
		}
	}
}