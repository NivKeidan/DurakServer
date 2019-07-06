package game_test

import (
	"DurakGo/game"
	"testing"
)

func TestGetKindCode(t *testing.T) {
	// Invalid strings
	stringsToTest := []string{"omg", "clubs", ""}
	for _, testString := range stringsToTest {
		k := game.Kind(testString)
		_, err := game.GetKindCode(k)
		if err == nil {
			t.Errorf("Kind %s was approved!\n", testString)
		}
	}

	// Valid strings
	stringsToTest = []string{"Clubs", "Diamonds", "Hearts", "Spades"}
	for _, testString := range stringsToTest {
		k := game.Kind(testString)
		_, err := game.GetKindCode(k)
		if err != nil {
			t.Errorf("Kind %s was not OK!\n", testString)
		}
	}
}

func TestGetCardKindByCode(t *testing.T) {
	// Valid values

	codesToTest := []string{"C", "S", "H", "D"}
	for _, code := range codesToTest {
		_, err := game.GetCardKindByCode(code)
		if err != nil {
			t.Errorf("Kind code %s was not converted to kind successfully\n", code)
		}
	}

	// Invalid values

	codesToTest = []string{"c", "O", "1", ""}
	for _, code := range codesToTest {
		_, err := game.GetCardKindByCode(code)
		if err == nil {
			t.Errorf("Kind code %s was converted to kind successfully\n", code)
		}
	}
}