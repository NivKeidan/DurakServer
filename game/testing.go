package game

import (
	"math/rand"
	"time"
)

type cardData struct {
	k Kind
	v uint
}

func makeCardData(k string, v int) *cardData {
	return &cardData{Kind(k), uint(v)}
}

func makeCard(k string, v int) *Card {
	return &Card{Kind: Kind(k), Value: uint(v)}
}

type cardPair struct {
	attackingCard *Card
	defendingCard *Card
}

func GetRandomCard() *Card {
	return &Card{Kind: getRandomKind(), Value: getRandomCardValue()}
}

func getRandomKind() Kind {
	rand.Seed(time.Now().Unix())
	return Kinds[rand.Intn(len(Kinds))]

}

func getRandomCardValue() uint {
	rand.Seed(time.Now().Unix())
	min := 6
	max := 14
	return uint(rand.Intn(max - min) + min)
}

func getBoardWithAttackingCardsOnly(cards ...*Card) *Board {
	b := NewBoard()
	for _, c := range cards {
		b.AddAttackingCard(c)
	}
	return b
}

func getBoardWithDefendedCard(att *Card, def *Card) *Board {
	b := NewBoard()
	b.AddAttackingCard(att)
	if err := b.AddDefendingCard(att, def); err != nil {
		return nil
	}
	return b
}