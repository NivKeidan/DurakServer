package game

import (
	"fmt"
)

type Kind string

const (
	Clubs    = Kind("Clubs")
	Spades   = Kind("Spades")
	Hearts   = Kind("Hearts")
	Diamonds = Kind("Diamonds")
)

var Kinds    = []Kind{Clubs, Spades, Hearts, Diamonds}

func GetKindCode(kind Kind) byte {
	return kind[0]
}

func GetCardKindByCode(kindCode string) (Kind, error) {
	switch kindCode {
	case "C":
		return Clubs, nil
	case "S":
		return Spades, nil
	case "D":
		return Diamonds, nil
	case "H":
		return Hearts, nil
	default:
		return "", fmt.Errorf("no such kind code: %v", kindCode)
	}
}
