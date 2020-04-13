package splenda

import (
	"encoding/base64"
	"errors"
	"math/rand"
)

// Impl implements Splenda's game logic.
type Impl struct {
	db  *DB
	rng rng
}

type realrng struct{}

func (realrng) Intn(n int) int {
	return rand.Intn(n)
}

// NewImpl creates a new Impl.
func NewImpl(db *DB) *Impl {
	return &Impl{
		db:  db,
		rng: realrng{},
	}
}

// NewImplSeed creates a new impl with the given psuedorandom seed.
func NewImplSeed(db *DB, seed int64) *Impl {
	return &Impl{
		db:  db,
		rng: rand.New(rand.NewSource(seed)),
	}
}

// ListGames lists all the games that the given user is in.
func (i *Impl) ListGames(userID string) ([]*GameSummary, error) {
	games, err := i.db.ListGames(userID)
	if err != nil {
		return nil, err
	}

	ret := []*GameSummary{}
	for gameID, players := range games {
		ret = append(ret, &GameSummary{
			ID:      gameID,
			Players: players,
		})
	}
	return ret, nil
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
func (i *Impl) NewGame(userID string, players []string) (string, error) {
	if find(userID, players) == -1 {
		return "", errors.New("you must be one of the players")
	}
	if len(players) < 2 {
		return "", errors.New("need at least two players")
	}
	if len(players) > 4 {
		return "", errors.New("no more than four players")
	}
	if !unique(players) {
		return "", errors.New("players must be unique")
	}

	gameID := newID()

	tx, err := i.db.NewTX(gameID)
	if err != nil {
		return "", err
	}
	defer tx.Close()

	players = shuffle(players, i.rng)

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

	nobles := pickNobles(len(players)+1, i.rng)
	if err := tx.InsertNobles(nobles); err != nil {
		return "", err
	}

	t1 := shuffleCards(tier1, i.rng)
	t2 := shuffleCards(tier2, i.rng)
	t3 := shuffleCards(tier3, i.rng)

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

	table, err := getTable(tx)
	if err != nil {
		return nil, err
	}
	game.Table = table

	players, err := getPlayers(tx)
	if err != nil {
		return nil, err
	}
	game.Players = players

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return game, nil
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
		if m.State() != play {
			return "", "", errors.New("can't do that right now")
		}

		delta := map[string]int{}
		for _, c := range colors {
			delta[c] = 1
		}

		if err := m.EarnCoins(tx, delta, delta); err != nil {
			return "", "", err
		}

		return m.NextPlayer()
	})
}

func unique(strs []string) bool {
	set := map[string]struct{}{}
	for _, str := range strs {
		set[str] = struct{}{}
	}
	return len(set) == len(strs)
}

// Take2 takes two coins of the same color.
func (i *Impl) Take2(gameID string, userID string, color string) (string, error) {
	if !isNormalColor(color) {
		return "", errors.New("invalid coin color")
	}

	m := mover{gameID: gameID, userID: userID, db: i.db}
	return m.Move(func(tx *TX) (string, string, error) {
		if m.State() != play {
			return "", "", errors.New("can't do that right now")
		}

		limit := map[string]int{color: 4}
		delta := map[string]int{color: 2}

		if err := m.EarnCoins(tx, limit, delta); err != nil {
			return "", "", err
		}

		return m.NextPlayer()
	})
}

// Reserve reserves a card.
func (i *Impl) Reserve(gameID string, userID string, tier int, index int) (string, error) {
	m := mover{gameID: gameID, userID: userID, db: i.db}
	return m.Move(func(tx *TX) (string, string, error) {
		if m.State() != play {
			return "", "", errors.New("can't do that right now")
		}

		// Make sure the player does not already have too many reserved cards.
		_, reservedIDs, err := tx.GetPlayerCards(userID)
		if err != nil {
			return "", "", err
		}
		if len(reservedIDs) >= 3 {
			return "", "", errors.New("too many cards already reserved")
		}

		// Grab the ID of the card that's currently in that position.
		// TODO: Handle reserving directly off the top of the deck.
		card, err := m.GetCardID(tx, tier, index)
		if err != nil {
			return "", "", err
		}

		// Insert it to the player's hand.
		reserved := true
		if err := tx.InsertPlayerCard(userID, card, reserved); err != nil {
			return "", "", err
		}

		// Replace it on the board.
		if err := m.DealCard(tx, tier, index); err != nil {
			return "", "", err
		}

		// Give the player a wildcard coin if we can.
		delta := map[string]int{wild: 1}
		err = m.EarnCoins(tx, delta, delta)
		if err != nil && err != ErrInsufficientCoins {
			return "", "", err
		}

		return m.NextPlayer()
	})
}

// Buy buys a card.
func (i *Impl) Buy(gameID string, userID string, tier int, index int) (string, error) {
	m := mover{gameID: gameID, userID: userID, db: i.db}
	return m.Move(func(tx *TX) (string, string, error) {
		if m.State() != play {
			return "", "", errors.New("can't do that right now")
		}

		// Grab the card that's currently in that position.
		cardID, err := m.GetCardID(tx, tier, index)
		if err != nil {
			return "", "", err
		}
		card, ok := cardFromID(cardID)
		if !ok {
			return "", "", errors.New("bogus card ID")
		}

		// Get the current card count.
		cards, err := m.GetCardCounts(tx)
		if err != nil {
			return "", "", err
		}

		// Pay the cost of the card.
		if err := m.PayCost(tx, cards, card.cost); err != nil {
			return "", "", err
		}

		// Insert it to the player's hand.
		reserved := false
		if err := tx.InsertPlayerCard(userID, cardID, reserved); err != nil {
			return "", "", err
		}
		cards[card.color]++

		// Replace it on the board.
		if err := m.DealCard(tx, tier, index); err != nil {
			return "", "", err
		}

		return nextState(&m, tx, cards)
	})
}

//
// Helper functions.
//

// NextState returns the next state to transition to after a player buys a card.
func nextState(m *mover, tx *TX, cards map[string]int) (string, string, error) {
	// Does this player now have enough cards to pick a noble? If yes, give them
	// time to pick one.
	can, err := m.CanAffordNoble(tx, cards)
	if err != nil {
		return "", "", err
	}
	if can {
		return picknoble, m.userID, nil
	}

	// Is this the last player, and if so has someone won the game? If so
	// stop playing and let them know they won.
	over, err := isGameOver(tx)
	if err != nil {
		return "", "", err
	}
	if over {
		return gameover, m.userID, nil
	}

	// Otherwise just go to the next player's turn.
	return m.NextPlayer()
}

// IsGameOver returns true if the game is now over.
func isGameOver(tx *TX) (bool, error) {
	players, err := getPlayers(tx)
	if err != nil {
		return false, err
	}

	for _, p := range players {
		if p.Points >= 15 {
			return true, nil
		}
	}

	return false, nil
}

// GetTable gets information about the table.
func getTable(tx *TX) (*Table, error) {
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

	cards, err := getCards(tx)
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

// GetCards gets the cards currently on the table.
func getCards(tx *TX) ([][]*Card, error) {
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

// GetPlayers gets data about the players in this game.
func getPlayers(tx *TX) ([]*Player, error) {
	userIDs, err := tx.GetPlayers()
	if err != nil {
		return nil, err
	}

	players := []*Player{}

	for _, userID := range userIDs {
		player, err := getPlayer(tx, userID)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

// GetPlayer gets data about the given player.
func getPlayer(tx *TX, userID string) (*Player, error) {
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
	cards, err := partitionCards(cardIDs)
	if err != nil {
		return nil, err
	}
	reserved, err := ToCards(reservedIDs)
	if err != nil {
		return nil, err
	}

	points := score(nobles, cards)

	return &Player{
		ID:       userID,
		Coins:    coins,
		Nobles:   nobles,
		Cards:    cards,
		Reserved: reserved,
		Points:   points,
	}, nil
}

// Partition partitions a player's cards by color.
func partitionCards(ids []string) (map[string][]*Card, error) {
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

// Score calculates the current score for a player.
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
