package game

import (
	"errors"
	"fmt"
)

type Board struct {
	cardsOnBoard []*CardOnBoard
}

func NewBoard() *Board {
	board := Board{}
	board.EmptyBoard()
	return &board
}

func (this *Board) EmptyBoard() {
	// Removes all cards from board

	this.cardsOnBoard = make([]*CardOnBoard, 0)
}

func (this *Board) GetAllCardsOnBoard() []*CardOnBoard {
	// Returns all cards on board
	// Does not remove cards from board
	cards := make([]*CardOnBoard, 0)
	for _, cardOnBoard := range this.cardsOnBoard {
		cards = append(cards, cardOnBoard)
	}
	return cards
}

func (this *Board) GetAllCards() []*Card {
	// Returns all cards that are on the board
	// Does NOT remove cards from board
	cards := make([]*Card, 0)
	for _, cardOnBoard := range this.cardsOnBoard {
		cards = append(cards, cardOnBoard.attackingCard)
		if cardOnBoard.defendingCard != nil {
			cards = append(cards, cardOnBoard.defendingCard)
		}
	}
	return cards
}

func (this *Board) AreAllCardsDefended() bool {
	for _, cardOnBoard := range this.cardsOnBoard {
		if cardOnBoard.defendingCard == nil {
			return false
		}
	}
	return true
}

func (this *Board) AddAttackingCard(card *Card) {
	newCardOnBoard := CardOnBoard{attackingCard: card}
	this.cardsOnBoard = append(this.cardsOnBoard, &newCardOnBoard)
}

func (this *Board) DefendCard(attackingCard *Card, defendingCard *Card, kozerKind *Kind) error {
	// Defends a card
	// Validates attacking card is on board
	// Validates defending card can defend the attacking card

	for _, cardOnBoard := range this.cardsOnBoard {
		if cardOnBoard.attackingCard.Kind == attackingCard.Kind &&
			cardOnBoard.attackingCard.Value == attackingCard.Value {
			if defendingCard.canDefendCard(attackingCard, kozerKind) {
				cardOnBoard.defendingCard = defendingCard
				return nil
			} else {
				return fmt.Errorf("%v can not defend %v", defendingCard, attackingCard)
			}
		}
	}
	return errors.New("attacking card is not on board")


}

func (this *Board) GetUnAnsweredCard() []*Card {
	// Returns all unanswered cards on board
	// Does NOT remove them from board

	cards := make([]*Card, 0)
	for _, cardOnBoard := range this.cardsOnBoard {
		if cardOnBoard.defendingCard == nil {
			cards = append(cards, cardOnBoard.attackingCard)
		}
	}
	return cards
}

func (this *Board) isEmpty() bool {
	return len(this.cardsOnBoard) == 0
}

func (this *Board) isCardLimitReached(numOfCardsInHand int) bool {
	// Checks if over total card limit on board, or if player has enough cards to defend

	return len(this.cardsOnBoard) >= MaxCardsPerAttack || len(this.GetUnAnsweredCard()) >= numOfCardsInHand
}

func (this *Board) canCardBeAdded(card *Card) bool {
	for _, currentCard := range this.cardsOnBoard {
		if currentCard.attackingCard.Value == card.Value ||
			(currentCard.defendingCard != nil && currentCard.defendingCard.Value == card.Value) {
			return true
		}
	}

	return false
}