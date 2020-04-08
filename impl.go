package splenda

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// Impl implements Splenda's game logic.
type Impl struct {
	db *DB
}

// NewImpl creates a new Impl.
func NewImpl(db *DB) *Impl {
	return &Impl{
		db: db,
	}
}

// ListGames lists all the games that the given user is in.
func (i *Impl) ListGames(userID string) (map[string][]string, error) {
	return i.db.ListGames(userID)
}

func newID() string {
	bs := make([]byte, 16)
	if _, err := rand.Read(bs); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bs)
}

func numCoins(players []string) int {
	switch len(players) {
	case 2:
		return 4
	case 3:
		return 5
	case 4:
		return 7
	default:
		panic("weird number of players")
	}
}

// NewGame creates a new game.
func (i *Impl) NewGame(players []string) (string, error) {
	if len(players) < 2 {
		return "", errors.New("need at least two players")
	}
	if len(players) > 4 {
		return "", errors.New("no more than four players")
	}

	gameID := newID()

	tx, err := i.db.NewTX(gameID)
	if err != nil {
		return "", err
	}
	defer tx.Close()

	// Set up the game table itself.
	if err := tx.InsertGame(players[0]); err != nil {
		return "", err
	}

	nc := numCoins(players)
	coins := map[string]int{
		red:   nc,
		blue:  nc,
		green: nc,
		black: nc,
		white: nc,
		wild:  5,
	}
	if err := tx.InsertCoins(coins); err != nil {
		return "", err
	}

	nobles := pick(len(players) + 1)
	if err := tx.InsertNobles(nobles); err != nil {
		return "", err
	}

	t1 := shuffle(tier1)
	t2 := shuffle(tier2)
	t3 := shuffle(tier3)

	if err := tx.InsertCards(t1[:4], t2[:4], t3[:4]); err != nil {
		return "", err
	}
	if err := tx.InsertDecks(t1[4:], t2[4:], t3[4:]); err != nil {
		return "", err
	}

	// Set up the players.
	if err := tx.InsertPlayers(players); err != nil {
		return "", err
	}

	allcolors := []string{red, blue, green, black, white, wild}
	for _, player := range players {
		if err := tx.InsertPlayerCoins(player, allcolors); err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return gameID, nil
}

// GetGame gets the current state of a given game.
func (i *Impl) GetGame(gameID string, userID string, ts string) (*Game, error) {
	tx, err := i.db.NewTX(gameID)
	if err != nil {
		return nil, err
	}
	defer tx.Close()

	if !tx.IsPlaying(userID) {
		return nil, errors.New("no such game")
	}

	game, err := tx.GetGameBasics()
	if err != nil {
		return nil, err
	}

	// TODO: Do something different if ts is current?

	table, err := i.getTable(tx)
	if err != nil {
		return nil, err
	}
	game.Table = table

	players, err := i.getPlayers(tx)
	if err != nil {
		return nil, err
	}
	game.Players = players

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return game, nil
}

func (i *Impl) getTable(tx *TX) (*Table, error) {
	coins, err := tx.GetCoins()
	if err != nil {
		return nil, err
	}

	nobleIDs, err := tx.GetNobles()
	if err != nil {
		return nil, err
	}
	nobles, err := ToNobles(nobleIDs)
	if err != nil {
		return nil, err
	}

	cards, err := i.getCards(tx)
	if err != nil {
		return nil, err
	}

	decks, err := tx.GetDecks()
	if err != nil {
		return nil, err
	}

	return &Table{
		Coins:  coins,
		Nobles: nobles,
		Cards:  cards,
		Decks:  decks,
	}, nil
}

func (i *Impl) getCards(tx *TX) ([][]*Card, error) {
	cards := [][]*Card{}
	for tier := 1; tier <= 3; tier++ {
		ids, err := tx.GetCards(tier)
		if err != nil {
			return nil, err
		}

		row, err := ToCards(ids)
		if err != nil {
			return nil, err
		}

		cards = append(cards, row)
	}
	return cards, nil
}

func (i *Impl) getPlayers(tx *TX) ([]*Player, error) {
	userIDs, err := tx.GetPlayers()
	if err != nil {
		return nil, err
	}

	players := []*Player{}

	for _, userID := range userIDs {
		coins, err := tx.GetPlayerCoins(userID)
		if err != nil {
			return nil, err
		}

		nobleIDs, err := tx.GetPlayerNobles(userID)
		if err != nil {
			return nil, err
		}
		nobles, err := ToNobles(nobleIDs)
		if err != nil {
			return nil, err
		}

		cardIDs, reservedIDs, err := tx.GetPlayerCards(userID)
		if err != nil {
			return nil, err
		}
		cards, err := partition(cardIDs)
		if err != nil {
			return nil, err
		}
		reserved, err := ToCards(reservedIDs)
		if err != nil {
			return nil, err
		}

		points := score(nobles, cards)

		players = append(players, &Player{
			ID:       userID,
			Coins:    coins,
			Nobles:   nobles,
			Cards:    cards,
			Reserved: reserved,
			Points:   points,
		})
	}

	return players, nil
}

func partition(ids []string) (map[string][]*Card, error) {
	ret := map[string][]*Card{}
	for _, id := range ids {
		card, err := ToCard(id)
		if err != nil {
			return nil, err
		}

		ret[card.Color] = append(ret[card.Color], card)
	}
	return ret, nil
}

func score(nobles []*Noble, cards map[string][]*Card) int {
	points := 0

	for _, noble := range nobles {
		points += noble.Points
	}

	for _, row := range cards {
		for _, card := range row {
			points += card.Points
		}
	}

	return points
}

// DeleteGame deletes a game.
func (i *Impl) DeleteGame(gameID string, userID string) error {
	tx, err := i.db.NewTX(gameID)
	if err != nil {
		return err
	}
	defer tx.Close()

	if !tx.IsPlaying(userID) {
		return errors.New("no such game")
	}

	if err := tx.DeleteGame(); err != nil {
		return err
	}

	return tx.Commit()
}

func unique(strs []string) bool {
	set := map[string]struct{}{}
	for _, str := range strs {
		set[str] = struct{}{}
	}
	return len(set) == len(strs)
}

// Take3 takes three coins of different colors.
func (i *Impl) Take3(gameID string, userID string, colors []string) (string, error) {
	if len(colors) != 3 {
		return "", errors.New("must specify three colors")
	}
	for _, c := range colors {
		if !isNormalColor(c) {
			return "", errors.New("invalid coin color")
		}
	}
	if !unique(colors) {
		return "", errors.New("colors must be unique")
	}

	m := mover{gameID: gameID, userID: userID, db: i.db}
	return m.Move(func(tx *TX) (string, string, error) {
		if m.game.State != "play" {
			return "", "", errors.New("can't do that right now")
		}

		delta := map[string]int{}
		for _, c := range colors {
			delta[c] = 1
		}

		if err := transferCoins(tx, userID, delta, delta); err != nil {
			return "", "", err
		}

		return nextPlayer(m.players, m.index)
	})
}

// Take2 takes two coins of the same color.
func (i *Impl) Take2(gameID string, userID string, color string) (string, error) {
	if !isNormalColor(color) {
		return "", errors.New("invalid coin color")
	}

	m := mover{gameID: gameID, userID: userID, db: i.db}
	return m.Move(func(tx *TX) (string, string, error) {
		if m.game.State != play {
			return "", "", errors.New("can't do that right now")
		}

		limit := map[string]int{color: 4}
		delta := map[string]int{color: 2}
		if err := transferCoins(tx, userID, limit, delta); err != nil {
			return "", "", err
		}

		return nextPlayer(m.players, m.index)
	})
}

// Reserve reserves a card.
func (i *Impl) Reserve(gameID string, userID string, tier int, index int) (string, error) {
	m := mover{gameID: gameID, userID: userID, db: i.db}
	return m.Move(func(tx *TX) (string, string, error) {
		if m.game.State != "play" {
			return "", "", errors.New("can't do that right now")
		}

		// Make sure the player does not already have too many reserved cards.
		_, reserved, err := tx.GetPlayerCards(userID)
		if err != nil {
			return "", "", err
		}
		if len(reserved) >= 3 {
			return "", "", errors.New("too many cards already reserved")
		}

		// Grab the card that's currently in that position.
		// TODO: Handle reserving directly off the top of the deck?
		cards, err := tx.GetCards(tier)
		if err != nil {
			return "", "", err
		}
		card := cards[index]
		if card == "" {
			return "", "", errors.New("no card there")
		}

		// Remove it from the board and replace with the next one from the deck.
		newcard, err := tx.GetTopCard(tier)
		if err != nil {
			return "", "", err
		}
		if newcard != "" {
			err = tx.TransferCard(tier, index, newcard)
		} else {
			err = tx.RemoveCard(tier, index)
		}
		if err != nil {
			return "", "", err
		}

		// Insert it to the player's hand.
		if err := tx.InsertPlayerCard(userID, card, true); err != nil {
			return "", "", err
		}

		// Give the player a wildcard coin if we can.
		delta := map[string]int{wild: 1}
		err = transferCoins(tx, userID, delta, delta)
		if err != nil && err != ErrInsufficientCoins {
			return "", "", err
		}

		return nextPlayer(m.players, m.index)
	})
}

// Buy buys a card.
func (i *Impl) Buy(gameID string, userID string, tier int, index int) (string, error) {
	m := mover{gameID: gameID, userID: userID, db: i.db}
	return m.Move(func(tx *TX) (string, string, error) {
		if m.game.State != "play" {
			return "", "", errors.New("can't do that right now")
		}

		// Grab the card that's currently in that position.
		cards, err := tx.GetCards(tier)
		if err != nil {
			return "", "", err
		}
		cardID := cards[index]
		if cardID == "" {
			return "", "", errors.New("no card there")
		}
		card, ok := cardFromID(cardID)
		if !ok {
			return "", "", errors.New("bogus card ID")
		}

		// Pay the cost of the card.
		cardIDs, _, err := tx.GetPlayerCards(userID)
		if err != nil {
			return "", "", err
		}
		pcards, err := partitionCounts(cardIDs)
		if err != nil {
			return "", "", err
		}
		if err := payCost(tx, userID, pcards, card.cost); err != nil {
			return "", "", err
		}

		// Remove it from the board and replace with the next one from the deck.
		newcard, err := tx.GetTopCard(tier)
		if err != nil {
			return "", "", err
		}
		if newcard != "" {
			err = tx.TransferCard(tier, index, newcard)
		} else {
			err = tx.RemoveCard(tier, index)
		}
		if err != nil {
			return "", "", err
		}

		// Insert it to the player's hand.
		if err := tx.InsertPlayerCard(userID, cardID, false); err != nil {
			return "", "", err
		}

		// TODO: does this player now have enough cards to pick a noble?
		// TODO: is this the last player, and if so has someone won the game?
		a syntax error is here

		return nextPlayer(m.players, m.index)
	})
}

func partitionCounts(ids []string) (map[string]int, error) {
	ret := map[string]int{}
	for _, id := range ids {
		card, ok := cardFromID(id)
		if !ok {
			return nil, errors.New("bogus card id")
		}
		ret[card.color]++
	}
	return ret, nil
}

func transferCoins(tx *TX, userID string, limits, coins map[string]int) error {
	bank, err := tx.GetCoins()
	if err != nil {
		return err
	}
	purse, err := tx.GetPlayerCoins(userID)
	if err != nil {
		return err
	}

	for color, limit := range limits {
		if bank[color] < limit {
			return ErrInsufficientCoins
		}
	}

	bankdiff := map[string]int{}
	pursediff := map[string]int{}

	for color, count := range coins {
		bankdiff[color] = bank[color] - count
		pursediff[color] = purse[color] + count
	}

	if err := tx.UpdateCoins(bankdiff); err != nil {
		return err
	}
	if err := tx.UpdatePlayerCoins(userID, pursediff); err != nil {
		return err
	}

	return nil
}

func payCost(tx *TX, userID string, cards map[string]int, cost map[string]int) error {
	bank, err := tx.GetCoins()
	if err != nil {
		return err
	}
	purse, err := tx.GetPlayerCoins(userID)
	if err != nil {
		return err
	}

	bankdiff := map[string]int{}
	pursediff := map[string]int{}
	wildsneeded := 0

	for color, needed := range cost {
		needed -= cards[color]
		if needed <= 0 {
			continue
		}

		remaining := purse[color] - needed
		if remaining >= 0 {
			// We can directly afford it.
			pursediff[color] = remaining
			bankdiff[color] = bank[color] + needed
		} else {
			// Keep track of the difference, which we can make up with wilds.
			pursediff[color] = 0
			bankdiff[color] += purse[color]

			wildsneeded -= remaining
		}
	}

	if wildsneeded > 0 {
		if wildsneeded > purse[wild] {
			return ErrInsufficientCoins
		}
		pursediff[wild] = purse[wild] - wildsneeded
		bankdiff[wild] = bank[wild] + wildsneeded
	}

	if err := tx.UpdateCoins(bankdiff); err != nil {
		return err
	}
	if err := tx.UpdatePlayerCoins(userID, pursediff); err != nil {
		return err
	}

	return nil
}

func nextPlayer(players []string, index int) (string, string, error) {
	next := index + 1
	if next >= len(players) {
		next = 0
	}
	return play, players[next], nil
}
