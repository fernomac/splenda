package splenda

import "database/sql"

// TX is a single DB transaction on a specific game.
type TX struct {
	db        *sql.DB
	tx        *sql.Tx
	gameID    string
	committed bool
}

//
// Query Methods.
//

// IsPlaying returns true if the given player is playing in this game.
func (t *TX) IsPlaying(userID string) bool {
	q := "SELECT 1 FROM players WHERE game_id = $1 AND user_id = $2"
	row := t.tx.QueryRow(q, t.gameID, userID)

	var ignored int
	err := row.Scan(&ignored)
	return err == nil
}

// GetGameBasics returns the basic info about a game.
func (t *TX) GetGameBasics() (*Game, error) {
	q := "SELECT ts, state, current FROM games WHERE id = $1"
	row := t.tx.QueryRow(q, t.gameID)

	var ts, state, current string
	if err := row.Scan(&ts, &state, &current); err != nil {
		return nil, err
	}

	return &Game{
		ID:      t.gameID,
		TS:      ts,
		State:   state,
		Current: current,
	}, nil
}

// GetCoins returns the number of coins of each color on the table.
func (t *TX) GetCoins() (map[string]int, error) {
	q := "SELECT color, count FROM game_coins WHERE game_id = $1"
	rows, err := t.tx.Query(q, t.gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	coins := map[string]int{}

	for rows.Next() {
		var color string
		var count int
		if err := rows.Scan(&color, &count); err != nil {
			return nil, err
		}
		coins[color] = count
	}

	return coins, rows.Err()
}

// GetNobles returns the IDs of the nobles currently on the table.
func (t *TX) GetNobles() ([]string, error) {
	q := "SELECT noble_id FROM game_nobles WHERE game_id = $1 ORDER BY INDEX ASC"
	rows, err := t.tx.Query(q, t.gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nobles := []string{}

	for rows.Next() {
		var noble string
		if err := rows.Scan(&noble); err != nil {
			return nil, err
		}
		nobles = append(nobles, noble)
	}

	return nobles, rows.Err()
}

// GetCards returns the IDs of the cards from the given tier currently on the table.
func (t *TX) GetCards(tier int) ([]string, error) {
	q := "SELECT index, card_id FROM game_cards WHERE game_id = $1 AND tier = $2"
	rows, err := t.tx.Query(q, t.gameID, tier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := [4]string{}

	for rows.Next() {
		var index int
		var card string
		if err := rows.Scan(&index, &card); err != nil {
			return nil, err
		}
		cards[index] = card
	}

	return cards[:], rows.Err()
}

// GetDecks gets the sizes of the decks for each tier.
func (t *TX) GetDecks() ([]int, error) {
	q := "SELECT tier, COUNT(*) FROM game_decks WHERE game_id = $1 GROUP BY tier"
	rows, err := t.tx.Query(q, t.gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := [3]int{0, 0, 0}

	for rows.Next() {
		var tier, count int
		if err := rows.Scan(&tier, &count); err != nil {
			return nil, err
		}
		counts[tier-1] = count
	}

	return counts[:], rows.Err()
}

// GetTopCard gets the top card on the given deck.
func (t *TX) GetTopCard(tier int) (string, error) {
	q := "SELECT card_id FROM game_decks WHERE game_id = $1 AND tier = $2 ORDER BY index ASC LIMIT 1"
	rows, err := t.tx.Query(q, t.gameID, tier)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", nil
	}

	var id string
	if err := rows.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

// GetPlayers returns the IDs of the players in the game.
func (t *TX) GetPlayers() ([]string, error) {
	q := "SELECT user_id FROM players WHERE game_id = $1 ORDER BY index ASC"
	rows, err := t.tx.Query(q, t.gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []string{}

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetPlayerCoins returns the number of coins of each color that the given player has.
func (t *TX) GetPlayerCoins(userID string) (map[string]int, error) {
	q := "SELECT color, count FROM player_coins WHERE game_id = $1 AND user_id = $2"
	rows, err := t.tx.Query(q, t.gameID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	coins := map[string]int{}

	for rows.Next() {
		var color string
		var count int
		if err := rows.Scan(&color, &count); err != nil {
			return nil, err
		}
		coins[color] = count
	}

	return coins, rows.Err()
}

// GetPlayerNobles returns the IDs of the nobles the given player has.
func (t *TX) GetPlayerNobles(userID string) ([]string, error) {
	q := "SELECT noble_id FROM player_nobles WHERE game_id = $1 AND user_id = $2"
	rows, err := t.tx.Query(q, t.gameID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []string{}

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GetPlayerCards returns the IDs of the cards the given player has.
func (t *TX) GetPlayerCards(userID string) ([]string, []string, error) {
	q := "SELECT card_id, reserved FROM player_cards WHERE game_id = $1 AND user_id = $2"
	rows, err := t.tx.Query(q, t.gameID, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	ids := []string{}
	reserved := []string{}

	for rows.Next() {
		var id string
		var res bool
		if err := rows.Scan(&id, &res); err != nil {
			return nil, nil, err
		}

		if res {
			reserved = append(reserved, id)
		} else {
			ids = append(ids, id)
		}
	}

	return ids, reserved, nil
}

//
// Insert Methods.
//

// InsertGame inserts a new game record with the given first player.
func (t *TX) InsertGame(firstPlayer string) error {
	q := "INSERT INTO games (id, ts, state, current) VALUES ($1, $2, $3, $4)"
	_, err := t.tx.Exec(q, t.gameID, 0, "play", firstPlayer)
	return err
}

// InsertCoins inserts the given initial coin records.
func (t *TX) InsertCoins(coins map[string]int) error {
	q := "INSERT INTO game_coins (game_id, color, count) VALUES ($1, $2, $3)"
	for color, count := range coins {
		if _, err := t.tx.Exec(q, t.gameID, color, count); err != nil {
			return err
		}
	}
	return nil
}

// InsertNobles inserts the given initial noble IDs.
func (t *TX) InsertNobles(nobles []string) error {
	q := "INSERT INTO game_nobles (game_id, index, noble_id) VALUES ($1, $2, $3)"
	for i, noble := range nobles {
		if _, err := t.tx.Exec(q, t.gameID, i, noble); err != nil {
			return err
		}
	}
	return nil
}

// InsertCards inserts the given set of cards for each tier.
func (t *TX) InsertCards(t1 []string, t2 []string, t3 []string) error {
	if err := t.doInsertCards(1, t1); err != nil {
		return err
	}
	if err := t.doInsertCards(2, t2); err != nil {
		return err
	}
	if err := t.doInsertCards(3, t3); err != nil {
		return err
	}
	return nil
}

func (t *TX) doInsertCards(tier int, cards []string) error {
	q := "INSERT INTO game_cards (game_id, tier, index, card_id) VALUES ($1, $2, $3, $4)"
	for i, card := range cards {
		if _, err := t.tx.Exec(q, t.gameID, tier, i, card); err != nil {
			return err
		}
	}
	return nil
}

// InsertDecks inserts the given decks for each tier.
func (t *TX) InsertDecks(d1 []string, d2 []string, d3 []string) error {
	if err := t.doInsertDecks(1, d1); err != nil {
		return err
	}
	if err := t.doInsertDecks(2, d2); err != nil {
		return err
	}
	if err := t.doInsertDecks(3, d3); err != nil {
		return err
	}
	return nil
}

func (t *TX) doInsertDecks(tier int, cards []string) error {
	q := "INSERT INTO game_decks (game_id, tier, index, card_id) VALUES ($1, $2, $3, $4)"
	for i, card := range cards {
		if _, err := t.tx.Exec(q, t.gameID, tier, i, card); err != nil {
			return err
		}
	}
	return nil
}

// InsertPlayers inserts initial player records for each of the given users.
func (t *TX) InsertPlayers(userIDs []string) error {
	q := "INSERT INTO players (game_id, user_id, index) VALUES ($1, $2, $3)"
	for i, userID := range userIDs {
		if _, err := t.tx.Exec(q, t.gameID, userID, i); err != nil {
			return err
		}
	}
	return nil
}

// InsertPlayerCoins inserts zeros for the given player's coin balance.
func (t *TX) InsertPlayerCoins(userID string, colors []string) error {
	q := "INSERT INTO player_coins (game_id, user_id, color, count) VALUES ($1, $2, $3, $4)"
	for _, color := range colors {
		if _, err := t.tx.Exec(q, t.gameID, userID, color, 0); err != nil {
			return err
		}
	}
	return nil
}

// InsertPlayerCard inserts a card into the player's hand.
func (t *TX) InsertPlayerCard(userID string, cardID string, reserved bool) error {
	q := "INSERT INTO player_cards (game_id, user_id, card_id, reserved) VALUES ($1, $2, $3, $4)"
	_, err := t.tx.Exec(q, t.gameID, userID, cardID, reserved)
	return err
}

//
// Update Methods.
//

// UpdateCoins updates the given coin records.
func (t *TX) UpdateCoins(coins map[string]int) error {
	q := "UPDATE game_coins SET count = $1 WHERE game_id = $2 AND color = $3"
	for color, count := range coins {
		if _, err := t.tx.Exec(q, count, t.gameID, color); err != nil {
			return err
		}
	}
	return nil
}

// UpdatePlayerCoins updates the given coin records.
func (t *TX) UpdatePlayerCoins(userID string, coins map[string]int) error {
	q := "UPDATE player_coins SET count = $1 WHERE game_id = $2 AND user_id = $3 AND color = $4"
	for color, count := range coins {
		if _, err := t.tx.Exec(q, count, t.gameID, userID, color); err != nil {
			return err
		}
	}
	return nil
}

// UpdatePlayerCard updates a card in the player's hand.
func (t *TX) UpdatePlayerCard(userID string, cardID string, reserved bool) error {
	q := "UPDATE player_cards SET reserved = $1 WHERE game_id = $2 AND user_id = $3 AND card_id = $4"
	_, err := t.tx.Exec(q, reserved, t.gameID, userID, cardID)
	if err != nil {
		panic(err)
	}
	return err
}

// UpdateGame updates the game state after a move.
func (t *TX) UpdateGame(curTS string, newstate string, newcurrent string) (string, error) {
	q := "UPDATE games SET ts = ts+1, state = $1, current = $2 WHERE id = $3 AND ts = $4 RETURNING ts"
	row := t.tx.QueryRow(q, newstate, newcurrent, t.gameID, curTS)

	var newts string
	if err := row.Scan(&newts); err != nil {
		return "", err
	}

	return newts, nil
}

// TransferCard transfers a card from a deck to the table.
func (t *TX) TransferCard(tier int, index int, cardID string) error {
	q := "DELETE FROM game_decks WHERE game_id = $1 AND tier = $2 AND card_id = $3"
	if _, err := t.tx.Exec(q, t.gameID, tier, cardID); err != nil {
		return err
	}

	q = "UPDATE game_cards SET card_id = $1 WHERE game_id = $2 AND tier = $3 AND index = $4"
	_, err := t.tx.Exec(q, cardID, t.gameID, tier, index)
	return err
}

//
// Delete Methods.
//

// DeleteCard removes a card from the board when the corresponding deck is empty.
func (t *TX) DeleteCard(tier int, index int) error {
	q := "DELETE FROM game_cards WHERE game_id = $1 AND tier = $2 AND index = $3"
	_, err := t.tx.Exec(q, t.gameID, tier, index)
	return err
}

// DeleteGame deletes a game record.
func (t *TX) DeleteGame() error {
	q := "DELETE FROM games WHERE id = $1"
	_, err := t.tx.Exec(q, t.gameID)
	return err
}

// Commit commits the current transaction.
func (t *TX) Commit() error {
	if err := t.tx.Commit(); err != nil {
		return err
	}
	t.committed = true
	return nil
}

// Close closes the current transaction, rolling back if not committed.
func (t *TX) Close() {
	if !t.committed {
		t.tx.Rollback()
	}
	t.db.Close()
}
