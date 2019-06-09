package game

import "encoding/json"

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
