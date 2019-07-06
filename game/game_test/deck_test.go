package game

import (
	"DurakGo/game"
	"testing"
)

func TestNewDeck(t *testing.T) {
	if _, err := game.NewDeck(); err != nil {
		t.Errorf("Failed creating new deck. Error: %s\n", err.Error())
	}
}

func TestShuffle(t *testing.T) {
	d, _ := game.NewDeck()
	d.Shuffle()
}