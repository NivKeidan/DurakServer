package game

import (
	"errors"
	"fmt"
)

type Player struct {
	cards []*Card
	Name string
	IsPlaying bool
	NextPlayer *Player
}

func NewPlayer(name string) *Player {
	return &Player{cards: make([]*Card, 0), Name: name}
}

func (this *Player) TakeCards(cards ...*Card) {
	// Adds all argument cards to hand
	this.cards = append(this.cards, cards...)
}

func (this *Player) GetCard(card *Card) (*Card, error) {
	// Gets a specific card and removes card from hand
	for i, currentCard := range this.cards {
		if currentCard.Value == card.Value && currentCard.Kind == card.Kind {
			this.cards = append(this.cards[:i], this.cards[i+1:]...)
			return currentCard, nil
		}
	}
	return nil, errors.New("no such card in player's hand")
}

func (this *Player) PeekCards() []*Card {
	// Returns all cards
	// Does NOT remove them from hand
	return this.cards
}

func (this *Player) GetNumOfCardsInHand() int {
	return len(this.cards)
}

func (this *Player) String() string {
	return fmt.Sprintf("%v: %v", this.Name, this.cards)
}