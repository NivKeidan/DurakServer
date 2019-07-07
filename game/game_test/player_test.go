package game

import (
	"DurakGo/game"
	"testing"
)

func TestNewPlayer(t *testing.T) {
	game.NewPlayer("exampleName")
}

func TestGetCard(t *testing.T) {
	c1 := makeCard("Hearts", 13)
	c2 := makeCard("Clubs", 6)
	cards := []*game.Card{c1, c2}
	p := game.NewPlayer("test")
	p.TakeCards(cards...)

	c1FromP, err := p.GetCard(c1)
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}
	if c1FromP != c1 {
		t.Errorf("Returned card is not the same\n")
	}

	c2FromP, err := p.GetCard(c2)
	if err != nil {
		t.Errorf("Error occurred: %s\n", err.Error())
	}
	if c2FromP != c2 {
		t.Errorf("Returned card is not the same\n")
	}

}

