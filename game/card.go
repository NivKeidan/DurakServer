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

func NewCard(kind Kind, value uint) (*Card, error) {
	card := &Card{Kind: kind, Value: value}
	if value < MinCardValue || value > MaxCardValue {
		return nil, errors.New("card value incorrect")
	}
	for _, k := range Kinds {
		if k == kind {
			return card, nil
		}
	}
	return nil, errors.New("kind " + string(kind) + " is not valid\n")
}

func NewCardByCode(code string) (*Card, error) {
	if code == "" {
		return nil, errors.New("no card returning nil")
	}
	kindCode := code[len(code)-1]
	valueCode := code[:len(code)-1]

	kind, err := GetCardKindByCode(string(kindCode))
	if err != nil { return nil, err }

	value, err := getCardValueByCode(valueCode)
	if err != nil { return nil, err }

	if newCard, err := NewCard(kind, value); err != nil {
		return nil, err
	} else {
		return newCard, nil
	}
}

func CardToCode(card *Card) (string, error) {
	valueCode, err := valueToCode(card.Value)
	if err != nil {
		return "", err
	} else {
		kindCodeByte, err := GetKindCode(card.Kind)
		if err != nil {
			return "", err
		}
		kindCode := string(kindCodeByte)
		return valueCode + kindCode, nil
	}
}

func (this *Card) CanDefendCard(attackCard *Card, kozerKind *Kind) bool {
	if this.isSameSuit(attackCard) {
		return this.Value > attackCard.Value
	} else {
		return this.Kind == *kozerKind
	}
}

func valueToCode(value uint) (string, error) {
	if value >= MinCardValue && value <= 10 {
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

func (this *Card) isSameSuit(card *Card) bool {
	return this.Kind == card.Kind
}

func getCardValueByCode(valueCode string) (uint, error) {
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
		if string(valueCode[0]) == "0" {
			return 0, errors.New("value has 0 in the beginning")
		}
		value, err := strconv.Atoi(valueCode)
		if err != nil { return 0, err}
		if value < MinCardValue || value > MaxCardValue {
			return 0, fmt.Errorf("bad card value: %v", value)
		}
		return uint(value), nil
	}
}

// JSON Serialization Override

func (this *Card) MarshalJSON() ([]byte, error) {
	code, err := CardToCode(this)
	if err != nil {return nil, err}
	return json.Marshal(code)
}

func (this *Card) UnmarshalJSON(data []byte) error {
	code := data[1:len(data)-1]
	kindCode := code[len(code)-1]
	valueCode := code[:len(code)-1]

	if kind, err := GetCardKindByCode(string(kindCode)); err != nil {
		return err
	} else {
		this.Kind = kind
	}

	if value, err := getCardValueByCode(string(valueCode)); err != nil {
		return err
	} else {
		this.Value = value
	}

	return nil
}

// Print override

func (this *Card) String() string {
	var valueString string

	switch int(this.Value) {
		case 11:
			valueString = "Jack"
		case 12:
			valueString = "Queen"
		case 13:
			valueString = "King"
		case 14:
			valueString = "Ace"
		default:
			valueString = fmt.Sprint(this.Value)
		}

	return fmt.Sprintf("%s of %s", valueString, this.Kind)
}
