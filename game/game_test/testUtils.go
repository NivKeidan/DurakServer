package game

import (
	"DurakGo/game"
	"math/rand"
	"time"
)

type cardData struct {
	k game.Kind
	v uint
}

func makeCardData(k string, v int) *cardData {
	return &cardData{game.Kind(k), uint(v)}
}

func makeCard(k string, v int) *game.Card {
	return &game.Card{Kind: game.Kind(k), Value: uint(v)}
}

type cardPair struct {
	attackingCard *game.Card
	defendingCard *game.Card
}

func GetRandomCard() *game.Card {
	return &game.Card{Kind: getRandomKind(), Value: getRandomCardValue()}
}

func getRandomKind() game.Kind {
	rand.Seed(time.Now().Unix())
	return game.Kinds[rand.Intn(len(game.Kinds))]

}

func getRandomCardValue() uint {
	rand.Seed(time.Now().Unix())
	min := 2
	max := 14
	return uint(rand.Intn(max - min) + min)
}

func MakeCardOnBoard(att *game.Card, def *game.Card) *game.CardOnBoard{
	return game.NewCardOnBoard(att, def)
}