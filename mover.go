package splenda

import "errors"

// A mover is a utility that holds shared code across different types of moves.
type mover struct {
	gameID string
	userID string
	db     *DB

	game    *Game
	players []string
	index   int
}

// A movefunc implements the actual business logic of a move.
type movefunc func(*TX) (string, string, error)

// Move executes the overall workflow of a move transaction, calling out to
// a movefunc that implements the actual business logic.
func (m *mover) Move(move movefunc) (string, error) {
	tx, err := m.db.NewTX(m.gameID)
	if err != nil {
		return "", err
	}
	defer tx.Close()

	// Pre-move work.
	if err := m.premove(tx); err != nil {
		return "", err
	}

	// Run the actual move.
	newstate, newplayer, err := move(tx)
	if err != nil {
		return "", err
	}

	// Clean up with post-move work.
	return m.postmove(tx, newstate, newplayer)
}

// Premove does the common work to set up for a game move.
func (m *mover) premove(tx *TX) error {
	game, err := tx.GetGameBasics() // TODO: FOR UPDATE to lock?
	if err != nil {
		return err
	}
	m.game = game

	// Read the set of players for this game, find the index of
	// the calling player, and confirm that it's their turn to move.
	players, err := tx.GetPlayers()
	if err != nil {
		return err
	}
	m.players = players

	index := find(m.userID, players)
	if index == -1 {
		return errors.New("no such game")
	}
	m.index = index

	if m.userID != game.Current {
		return errors.New("not your turn")
	}

	return nil
}

func find(needle string, haystack []string) int {
	for i, hay := range haystack {
		if hay == needle {
			return i
		}
	}
	return -1
}

// Postmove does the common work to finish up after a move.
func (m *mover) postmove(tx *TX, newstate string, newplayer string) (string, error) {
	ts, err := tx.UpdateGame(m.game.TS, newstate, newplayer)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return ts, nil
}

// State returns the current state of the game.
func (m *mover) State() string {
	return m.game.State
}

// IsRoundOver returns true if this is the last turn for the round.
func (m *mover) IsRoundOver() bool {
	return m.index == len(m.players)-1
}

// GetCardID gets the ID of the card at a given index.
func (m *mover) GetCardID(tx *TX, tier int, index int) (string, error) {
	cards, err := tx.GetCards(tier)
	if err != nil {
		return "", err
	}
	card := cards[index]
	if card == "" {
		return "", errors.New("no card there")
	}
	return card, nil
}

// GetCardCounts gets the number of cards of each color that the player has.
func (m *mover) GetCardCounts(tx *TX) (map[string]int, error) {
	cards, _, err := tx.GetPlayerCards(m.userID)
	if err != nil {
		return nil, err
	}

	ret := map[string]int{}

	for _, id := range cards {
		card, ok := cardFromID(id)
		if !ok {
			return nil, errors.New("bogus card ID")
		}
		ret[card.color]++
	}

	return ret, nil
}

// EarnCoins transfers coins from the bank to the player if possible.
func (m *mover) EarnCoins(tx *TX, limits, coins map[string]int) error {
	bank, err := tx.GetCoins()
	if err != nil {
		return err
	}
	purse, err := tx.GetPlayerCoins(m.userID)
	if err != nil {
		return err
	}

	for color, limit := range limits {
		if bank[color] < limit {
			return ErrInsufficientCoins
		}
	}

	newbank := map[string]int{}
	newpurse := map[string]int{}

	for color, count := range coins {
		newbank[color] = bank[color] - count
		newpurse[color] = purse[color] + count
	}

	if err := tx.UpdateCoins(newbank); err != nil {
		return err
	}
	if err := tx.UpdatePlayerCoins(m.userID, newpurse); err != nil {
		return err
	}

	// TODO: If player has more than 10 coins now, make then give some back.

	return nil
}

// PayCost attempts to pay the cost for a given card.
func (m *mover) PayCost(tx *TX, cards, cost map[string]int) error {
	bank, err := tx.GetCoins()
	if err != nil {
		return err
	}
	purse, err := tx.GetPlayerCoins(m.userID)
	if err != nil {
		return err
	}

	newbank := map[string]int{}
	newpurse := map[string]int{}
	wildsneeded := 0

	for color, coins := range cost {
		coincost, wildcost := calculateActualCost(cards[color], purse[color], coins)
		if coincost > 0 {
			newbank[color] = bank[color] + coincost
			newpurse[color] = purse[color] - coincost
		}
		wildsneeded += wildcost
	}

	if wildsneeded > 0 {
		if wildsneeded > purse[wild] {
			return ErrInsufficientCoins
		}
		newbank[wild] = bank[wild] + wildsneeded
		newpurse[wild] = purse[wild] - wildsneeded
	}

	if err := tx.UpdateCoins(newbank); err != nil {
		return err
	}
	if err := tx.UpdatePlayerCoins(m.userID, newpurse); err != nil {
		return err
	}

	return nil
}

func calculateActualCost(cards, purse, cost int) (int, int) {
	cost -= cards
	if cost <= 0 {
		// We have enough cards to totally cancel the cost.
		return 0, 0
	}

	// We can directly afford it, pay the cost in regular coins.
	if cost <= purse {
		return cost, 0
	}

	// If we can't we need to spend some wilds.
	return purse, cost - purse
}

// DealCard deals a card from the deck onto the board.
func (m *mover) DealCard(tx *TX, tier int, index int) error {
	newcard, err := tx.GetTopCard(tier)
	if err != nil {
		return err
	}

	if newcard != "" {
		err = tx.TransferCard(tier, index, newcard)
	} else {
		err = tx.DeleteCard(tier, index)
	}
	if err != nil {
		return err
	}

	return nil
}

// CanAffordNoble checks if the player can now afford a noble.
func (m *mover) CanAffordNoble(tx *TX, cards map[string]int) (bool, error) {
	nobles, err := tx.GetNobles()
	if err != nil {
		return false, err
	}

	for _, id := range nobles {
		noble, ok := nobleFromID(id)
		if !ok {
			return false, errors.New("invalid noble ID")
		}

		if canAfford(cards, noble.cost) {
			return true, nil
		}
	}

	return false, nil
}

func canAfford(cards, cost map[string]int) bool {
	for color, num := range cost {
		if cards[color] < num {
			return false
		}
	}
	return true
}

// NextPlayer returns a triple that causes postmove to move to the next
// player's turn.
func (m *mover) NextPlayer() (string, string, error) {
	next := m.index + 1
	if next >= len(m.players) {
		next = 0
	}
	return play, m.players[next], nil
}
