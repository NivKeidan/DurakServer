package game

import (
	"errors"
	"fmt"
)

const CardsPerPlayer = 6
const MaxCardsPerAttack = 6

type Game struct {
	board           *Board
	deck            *Deck
	players         []*Player
	startingPlayer  *Player
	defendingPlayer *Player
	KozerCard       *Card
	numOfPlayers	int
}

// Server API

func NewGame(names ...string) (*Game, error) {
	// Create new deck
	deck, err := NewDeck()
	if err != nil { return nil, err}
	deck.Shuffle()

	// Create players
	players := make([]*Player, 0)
	var lastPlayer *Player
	for _, name := range names {
		player := NewPlayer(name)
		players = append(players, player)
		if lastPlayer != nil {
			lastPlayer.NextPlayer = player
		}
		lastPlayer = player
	}
	lastPlayer.NextPlayer = players[0]

	// Prepare game and cards
	game := Game{board: NewBoard(), deck: deck, players: players, numOfPlayers: len(names)}
	game.dealCards()
	game.chooseKozer()
	game.startGame()

	return &game, nil
}

func (this *Game) Attack(player *Player, card *Card) error {

	if !this.canPlayerAttackNow(player) {
		return fmt.Errorf("%s can not add attack now", player.Name)
	}

	if this.board.IsCardLimitReached(len(this.defendingPlayer.cards)) {
		return errors.New("card limit reached")
	}

	if !this.board.IsEmpty() && !this.board.CanCardBeAdded(card) {
		return fmt.Errorf("%s is not a valid card to attack with at this moment", card)
	}

	// Remove card from player
	card, err := player.GetCard(card)
	if err != nil {return err}

	this.board.AddAttackingCard(card)
	return nil

}

func (this *Game) Defend(player *Player, attackingCard *Card, defendingCard *Card) error {

	if this.defendingPlayer != player {
		return fmt.Errorf("%s is not defending now", player.Name)
	}

	// Remove card from player
	defendingCard, err := player.GetCard(defendingCard)

	if err != nil {
		return err
	}

	// Add card to board
	err = this.board.DefendCard(attackingCard, defendingCard, &this.KozerCard.Kind)

	if err != nil {
		player.TakeCards(defendingCard)  // Return card to player
		return err
	}
	return nil
}

func (this *Game) MoveToBita() error {
	if this.board.IsEmpty() {
		return errors.New("board is empty")
	}

	if !this.board.AreAllCardsDefended() {
		return errors.New("some cards are un defended")
	}
	this.finalizeTurn(true)
	return nil
}

func (this *Game) HandlePlayerTakesCard() {

	cards := this.board.PeekCards()
	this.defendingPlayer.TakeCards(cards...)
	this.finalizeTurn(false)
}

func (this *Game) IsGameOver() bool {
	return this.numOfPlayers < 2
}

func (this *Game) IsDraw() bool {
	return this.numOfPlayers == 0
}

func (this *Game) GetPlayerByName(name string) (*Player, error) {
	for _, player := range this.players {
		if player.Name == name {
			return player, nil
		}
	}
	return nil, fmt.Errorf("no such player exists: %s", name)
}

func (this *Game) NumOfCardsLeftInDeck() int {
	return this.deck.GetNumOfCardsLeft()
}

func (this *Game) GetCardsOnBoard() []*CardOnBoard {
	return this.board.PeekCardsOnBoard()
}

func (this *Game) GetNumOfCardsLeftInDeck() int {
	return this.deck.GetNumOfCardsLeft()
}

func (this *Game) GetLosingPlayer() *Player {
	if !this.IsGameOver() {
		return nil
	} else if this.IsDraw() {
		return nil
	}
	for _, p := range this.players {
		if p.GetNumOfCardsInHand() != 0 {
			return p
		}
	}
	return nil
}

func (this *Game) GetStartingPlayer() *Player {
	return this.startingPlayer
}

func (this *Game) GetDefendingPlayer() *Player {
	return this.defendingPlayer
}

func (this *Game) GetLosingPlayerName() string {
	losingPlayer := this.GetLosingPlayer()
	if losingPlayer == nil {
		return ""
	} else {
		return losingPlayer.Name
	}

}

func (this *Game) GetPlayersCardsMap() map[string][]*Card {
	playerCards := make(map[string][]*Card)
	for _, player := range this.players {
		cards := player.PeekCards()
		playerCards[player.Name] = cards
	}
	return playerCards

}

func (this *Game) GetPlayerNamesArray() []string {
	arr := make([]string, 0)
	for _, player := range this.players {
		arr = append(arr, player.Name)
	}
	return arr
}

// Internal methods

func (this *Game) finalizeTurn(wasDefendedSuccessfully bool) {

	// Removes cards from board, fills up players' hands

	this.board.EmptyBoard()
	this.fillUpCards()
	if this.deck.GetNumOfCardsLeft() == 0 {
		this.removePlayersThatFinished()
	}

	if !this.IsGameOver() {
		this.setUpNextTurn(wasDefendedSuccessfully)
	}
}

func (this *Game) dealCards() {
	for i := 1; i <= CardsPerPlayer; i++ {
		for _, player := range this.players {
			player.TakeCards(this.deck.GetNextCard())
		}
	}
}

func (this *Game) chooseKozer() {
	lastCardInDeck := this.deck.PeekLastCard()
	this.KozerCard = lastCardInDeck
}

func (this *Game) startGame() {
	this.startingPlayer = this.getStartingPlayer()
	this.defendingPlayer = this.startingPlayer.NextPlayer
}

func (this *Game) getStartingPlayer() *Player {
	// Check player with lowest kozer, or use default

	kozerKind := this.KozerCard.Kind
	minValue := uint(15)                    // Use value higher than highest value
	playerStarting := this.players[0] // Use first player, or any

	for _, player := range this.players {
		for _, card := range player.PeekCards() {
			if card.Kind == kozerKind && card.Value < uint(minValue) {
				playerStarting = player
				minValue = card.Value
			}
		}
	}

	// TODO Add attack durak from last round

	return playerStarting
}

func (this *Game) canPlayerAttackNow(player *Player) bool {
	// Checks if a player has the right to attack with a card

	if this.board.IsEmpty() {
		return this.startingPlayer == player
	} else {
		return player != this.defendingPlayer
	}
}

func (this *Game) fillUpCards() {

	// Check if there is a deck
	numOfCardsInDeck := this.deck.GetNumOfCardsLeft()
	if numOfCardsInDeck == 0 {
		return
	}

	playerFillingUpLast := this.getPlayerFillingUpLast()
	playerFillingUp := this.getPlayerFillingUpFirst()
	playerFilledUpCounter := 0

	for playerFilledUpCounter < this.numOfPlayers {
		if playerFilledUpCounter != this.numOfPlayers - 1 {
			if playerFillingUp == playerFillingUpLast {
				playerFillingUp = playerFillingUp.NextPlayer
			} else {
				this.fillUpCardsForPlayer(playerFillingUp)
				playerFillingUp = playerFillingUp.NextPlayer
				playerFilledUpCounter = playerFilledUpCounter + 1
			}
		} else {
			this.fillUpCardsForPlayer(playerFillingUpLast)
			break
		}
	}

}

func (this *Game) getPlayerFillingUpLast() *Player {
	return this.defendingPlayer
}

func (this *Game) getPlayerFillingUpFirst() *Player {
	return this.startingPlayer
}

func (this *Game) fillUpCardsForPlayer(player *Player) {

	for CardsPerPlayer- player.GetNumOfCardsInHand() > 0 {
		if this.deck.GetNumOfCardsLeft() == 0 {
			return
		}
		newCard := this.deck.GetNextCard()
		player.TakeCards(newCard)
	}
}

func (this *Game) setUpNextTurn(wasLastTurnDefended bool) {

	if wasLastTurnDefended && this.defendingPlayer.GetNumOfCardsInHand() > 0 {
			this.startingPlayer = this.defendingPlayer
	} else {
		this.startingPlayer = this.defendingPlayer.NextPlayer
	}

	this.defendingPlayer = this.startingPlayer.NextPlayer
}

func (this *Game) removePlayersThatFinished() {

	currentPlayer := this.defendingPlayer
	playersRemoved := 0
	for i := 0; i < this.numOfPlayers; i++ {
		if currentPlayer.GetNumOfCardsInHand() == 0 {
			playersRemoved++
			previousPlayer := this.getPreviousPlayer(currentPlayer)
			previousPlayer.NextPlayer = currentPlayer.NextPlayer
		}
		currentPlayer = currentPlayer.NextPlayer
	}
	this.numOfPlayers = this.numOfPlayers - playersRemoved
}

func (this *Game) getPreviousPlayer(player *Player) *Player {
	p := player
	for p.NextPlayer != player {
		p = p.NextPlayer
	}
	return p
}