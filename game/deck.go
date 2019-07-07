package game

import (
	"math/rand"
	"time"
)

type Deck struct {
	cards []*Card
}

func NewDeck() (*Deck, error) {
	deck := Deck{}
	deck.cards = make([]*Card, 0)
	for v := MinCardValue; v <= MaxCardValue; v++ {
		for _, kind := range Kinds {
			card, err := NewCard(kind, uint(v))
			if err == nil {
				deck.cards = append(deck.cards, card)
			} else { return nil, err}
		}
	}
	return &deck, nil
}

func (this *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(this.cards), func(i, j int) {
		this.cards[i], this.cards[j] = this.cards[j], this.cards[i]
	})
}

func (this *Deck) GetNextCard() *Card {
	// Gets and removes card from deck

	if len(this.cards) > 0 {
		card := this.cards[0]
		this.cards = this.cards[1:]
		return card
	} else {
		return nil
	}
}

func (this *Deck) PeekLastCard() *Card {
	// Does not remove card from deck (used for kozer card)

	return this.cards[len(this.cards)-1]
}

func (this *Deck) GetNumOfCardsLeft() int {
	return len(this.cards)
}