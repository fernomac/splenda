package splenda

import "errors"

type mover struct {
	gameID string
	userID string
	db     *DB

	game    *Game
	players []string
	index   int
}

type movefunc func(*TX) (string, string, error)

func (m *mover) Move(move movefunc) (string, error) {
	tx, err := m.db.NewTX(m.gameID)
	if err != nil {
		return "", err
	}
	defer tx.Close()

	if err := m.premove(tx); err != nil {
		return "", err
	}

	newstate, newplayer, err := move(tx)
	if err != nil {
		return "", err
	}

	return m.postmove(tx, newstate, newplayer)
}

func (m *mover) premove(tx *TX) error {
	// TODO: FOR UPDATE to lock?
	game, err := tx.GetGameBasics()
	if err != nil {
		return err
	}
	m.game = game

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
