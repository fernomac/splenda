package splenda

import "fmt"

// UserList is a list of users.
type UserList struct {
	Users []string `json:"users"`
}

// Login is a request to log in.
type Login struct {
	ID       string `json:"id"`
	Password string `json:"pw"`
}

// GameSummary lists the ID and players of a given game.
type GameSummary struct {
	ID      string   `json:"id"`
	Players []string `json:"players"`
}

// GameList lists the current games.
type GameList struct {
	Games []*GameSummary `json:"games,omitempty"`
}

// Noble describes a noble tile.
type Noble struct {
	ID     string         `json:"id"`
	Points int            `json:"points"`
	Cost   map[string]int `json:"cost"`
}

// ToNobles hydrates a list of Noble DTOs from their IDs.
func ToNobles(ids []string) ([]*Noble, error) {
	ret := []*Noble{}
	for _, id := range ids {
		if id == "" {
			ret = append(ret, nil)
			continue
		}

		noble, ok := nobleFromID(id)
		if !ok {
			return nil, fmt.Errorf("bogus noble id: %v", id)
		}

		ret = append(ret, &Noble{
			ID:     id,
			Points: 3, // All nobles are worth three points.
			Cost:   noble.cost,
		})
	}
	return ret, nil
}

// Card describes a gem card.
type Card struct {
	ID     string         `json:"id"`
	Color  string         `json:"color"`
	Points int            `json:"points"`
	Cost   map[string]int `json:"cost"`
}

// ToCards hydrates a list of Card DTOs from their IDs.
func ToCards(ids []string) ([]*Card, error) {
	ret := []*Card{}
	for _, id := range ids {
		card, err := ToCard(id)
		if err != nil {
			return nil, err
		}
		ret = append(ret, card)
	}
	return ret, nil
}

// ToCard hydrates a Card DTO from its ID.
func ToCard(id string) (*Card, error) {
	if id == "" {
		return nil, nil
	}

	card, ok := cardFromID(id)
	if !ok {
		return nil, fmt.Errorf("bogus card id: %v", id)
	}

	return &Card{
		ID:     id,
		Color:  card.color,
		Points: card.points,
		Cost:   card.cost,
	}, nil
}

// Table describes the state of the shared table.
type Table struct {
	Coins  map[string]int `json:"coins"`
	Nobles []*Noble       `json:"nobles"`
	Cards  [][]*Card      `json:"cards"`
	Decks  []int          `json:"decks"`
}

// Player describes the state of a player's hand.
type Player struct {
	ID       string             `json:"id"`
	Coins    map[string]int     `json:"coins"`
	Nobles   []*Noble           `json:"nobles"`
	Cards    map[string][]*Card `json:"cards"`
	Reserved []*Card            `json:"reserved"`
	Points   int                `json:"points"`
}

// Game describes the overall state of the game.
type Game struct {
	ID      string `json:"id"`
	TS      string `json:"ts"`
	State   string `json:"state"`
	Current string `json:"current"`

	Table   *Table    `json:"table"`
	Players []*Player `json:"players"`
}

// Take3 is a request to take three coins.
type Take3 struct {
	Colors []string `json:"colors"`
}

// Take2 is a request to take two coins.
type Take2 struct {
	Color string `json:"color"`
}

// Buy is a request to buy a card.
type Buy struct {
	Tier  int `json:"tier"`
	Index int `json:"index"`
}

// TS is a response containing an updated timestamp.
type TS struct {
	TS string `json:"ts"`
}
