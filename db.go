package splenda

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// DB implements Splenda's DB access layer.
type DB struct {
	url string
}

// NewDB returns a new DB.
func NewDB(url string) *DB {
	return &DB{
		url: url,
	}
}

// TODO: Connection pooling?
func (d *DB) open() (*sql.DB, error) {
	return sql.Open("postgres", d.url)
}

// ApplySchema applies a bunch of SQL statements.
func (d *DB) ApplySchema() error {
	db, err := d.open()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, stmt := range schema {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListUsers lists all the currently registered users. What could possibly go wrong?
func (d *DB) ListUsers() ([]string, error) {
	db, err := d.open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id FROM users")
	if err != nil {
		return nil, err
	}

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

// NewUser creates a new user in the DB.
func (d *DB) NewUser(userID string, hash string) error {
	db, err := d.open()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO users (id, hash) VALUES ($1, $2)", userID, hash)
	if err != nil {
		if pe, ok := err.(*pq.Error); ok {
			code := pe.Code.Name()
			if code == "unique_violation" {
				return errors.New("user already exists")
			}
		}
		return fmt.Errorf("database error: %v", err)
	}

	return nil
}

// GetUserHash gets the given user's password hash.
func (d *DB) GetUserHash(userID string) (string, error) {
	db, err := d.open()
	if err != nil {
		return "", err
	}
	defer db.Close()

	hash := ""
	row := db.QueryRow("SELECT hash FROM users WHERE id=$1", userID)
	if err := row.Scan(&hash); err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("no such user")
		}
		return "", err
	}

	return hash, nil
}

// ListGames lists the currently running games, and who is playing.
func (d *DB) ListGames(userID string) (map[string][]string, error) {
	db, err := d.open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	q := "CREATE TEMP TABLE games AS (SELECT game_id AS id FROM players WHERE user_id=$1)"
	if _, err := db.Exec(q, userID); err != nil {
		return nil, err
	}

	ret := map[string][]string{}

	q = "SELECT id, user_id FROM games, players WHERE id=game_id ORDER BY index ASC"
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, uid string
		if err := rows.Scan(&id, &uid); err != nil {
			return nil, err
		}

		ret[id] = append(ret[id], uid)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if _, err := db.Exec("DROP TABLE games"); err != nil {
		return nil, err
	}

	return ret, nil
}

// NewTX begins a new transaction on the given game.
func (d *DB) NewTX(gameID string) (*TX, error) {
	db, err := d.open()
	if err != nil {
		return nil, err
	}

	dbtx, err := db.Begin()
	if err != nil {
		db.Close()
		return nil, err
	}

	return &TX{
		db:     db,
		tx:     dbtx,
		gameID: gameID,
	}, nil
}
