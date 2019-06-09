package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)


type Card struct {
	Kind  Kind
	Value uint
}

func (this *Card) MarshalJSON() ([]byte, error) {
	code, err := cardToCode(this)
	if err != nil {return nil, err}
	return json.Marshal(code)
}  // JSON Serialization Override

func NewCard(kind Kind, value uint) (*Card, error) {
	card := Card{Kind: kind, Value: value}
	return &card, nil
}

func NewCardByCode(code string) (*Card, error) {
	kindCode := []byte(code)[len(code)-1]
	valueCode := code[:len(code)-1]

	kind, err := GetCardKindByCode(kindCode)
	if err != nil { return nil, err }

	value, err := GetCardValueByCode(valueCode)
	if err != nil { return nil, err }

	newCard := &Card{Kind: kind, Value: value}
	return newCard, nil
}

func GetCardValueByCode(valueCode string) (uint, error) {
	switch valueCode {
	case "A":
		return 14, nil
	case "K":
		return 13, nil
	case "Q":
		return 12, nil
	case "J":
		return 11, nil
	default:
		value, err := strconv.Atoi(valueCode)
		if err != nil { return 0, err}
		if value <= 0 || value > 10 {
			return 0, fmt.Errorf("bad card value: %v", value)
		}
		return uint(value), nil
	}
}

func cardToCode(card *Card) (string, error) {
	valueCode, err := valueToCode(card.Value)
	if err != nil {return "", err} else {
		kindCode := string(GetKindCode(card.Kind))
		return valueCode + kindCode, nil
	}
}

func valueToCode(value uint) (string, error) {
	if value > 0 && value <= 10 {
		return fmt.Sprint(value), nil
	} else {
		switch value {
		case 11:
			return "J", nil
		case 12:
			return "Q", nil
		case 13:
			return "K", nil
		case 14:
			return "A", nil
		default:
			return "", errors.New("no such card value")
		}
	}
}

func (this *Card) canDefendCard(card *Card, kozerKind *Kind) bool {
	if this.isSameSuit(card) {
		return this.Value > card.Value
	} else {
		return this.Kind == *kozerKind
	}
}

func (this *Card) isSameSuit(card *Card) bool {
	return this.Kind == card.Kind
}