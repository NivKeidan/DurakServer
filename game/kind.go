package game

import (
	"errors"
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

func GetKindCode(kind Kind) (byte, error) {
	for _, k := range Kinds {
		if kind == k {
			return kind[0], nil
		}
	}

	return 0, errors.New("unknown kind")
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
