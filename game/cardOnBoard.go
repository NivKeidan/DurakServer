package game

import (
	"encoding/json"
	"fmt"
)

type CardOnBoard struct {
	attackingCard *Card
	defendingCard *Card
}

func (this *CardOnBoard) MarshalJSON() ([]byte, error) {  // JSON Serialization override
	cardOnBoardArray := make([]string, 0)
	attackingCardCode, err := cardToCode(this.attackingCard)
	if err != nil {
		return nil, err
	} else {
		cardOnBoardArray = append(cardOnBoardArray, attackingCardCode)
	}

	if this.defendingCard != nil {
		defendingCardCode, err := cardToCode(this.defendingCard)
		if err != nil {return nil, err} else {
			cardOnBoardArray = append(cardOnBoardArray, defendingCardCode)
		}
	} else {
		cardOnBoardArray = append(cardOnBoardArray, "")
	}

	return json.Marshal(cardOnBoardArray)
}

func (this *CardOnBoard) UnmarshalJSON(data []byte) error {
	// Incoming data is (for example): ["7C",""]

	helperObj := make([]string, 0)
	if err := json.Unmarshal(data, &helperObj); err != nil {
		return fmt.Errorf("could not unmarshal data: %s\n", err)
	}

	if attackingCard, err := NewCardByCode(helperObj[0]); err != nil {
		return fmt.Errorf("could not parse attacking card code: %s\nError: %s\n", helperObj[0], err)
	} else {
		this.attackingCard = attackingCard
	}

	if defendingCard, err := NewCardByCode(helperObj[1]); err != nil {
		this.defendingCard = nil
	} else {
		this.defendingCard = defendingCard
	}
	return nil
}
