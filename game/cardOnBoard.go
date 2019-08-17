package game

import (
	"encoding/json"
	"fmt"
)

type CardOnBoard struct {
	attackingCard *Card
	attackingCardOwner *Player
	defendingCard *Card
	defendingCardOwner *Player
}

func (this *CardOnBoard) GetAttackingCard() *Card {
	return this.attackingCard
}

func (this *CardOnBoard) GetDefendingCard() *Card {
	return this.defendingCard
}

func (this *CardOnBoard) MarshalJSON() ([]byte, error) {  // JSON Serialization override
	cardOnBoardArray := make([]string, 0)
	attackingCardCode, err := CardToCode(this.attackingCard)
	if err != nil {
		return nil, err
	} else {
		cardOnBoardArray = append(cardOnBoardArray, attackingCardCode)
	}

	if this.defendingCard != nil {
		defendingCardCode, err := CardToCode(this.defendingCard)
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

func (this *CardOnBoard) String() string {
	var returnString string

	returnString = fmt.Sprintf("%v", this.attackingCard)
	if this.defendingCard == nil {
		returnString = fmt.Sprintf("%s (undefended)", returnString)
	} else {
		returnString = fmt.Sprintf("%s (defended by %v)", returnString, this.defendingCard)
	}

	return returnString
}
