package game

import (
	"testing"
)

func TestNewDeck(t *testing.T) {
	if _, err := NewDeck(); err != nil {
		t.Errorf("Failed creating new deck. Error: %s\n", err.Error())
	}
}

func TestShuffle(t *testing.T) {
	d, _ := NewDeck()
	d.Shuffle()
}