package splenda

import (
	"os"
	"os/exec"
	"testing"
)

func TestTwoPlayers(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	defer cleanup(url)

	impl, err := setup(url)
	if err != nil {
		t.Fatal(err)
	}

	id, err := impl.NewGame("user1", []string{"user1", "user2"})
	if err != nil {
		t.Fatal(err)
	}

	game, err := impl.GetGame(id, "user1", "0")
	if err != nil {
		t.Fatal(err)
	}
	assertGameState(t, game, id, play, "user2")
	assertCoins(t, game, map[string]int{red: 4, green: 4, blue: 4, black: 4, white: 4, wild: 5})
	assertNobles(t, game, []string{"macchiavelli", "charles_v", "catherine_of_medici"})
	assertCards(t, game.Table.Cards[2], []string{"3_7_3_0", "3_6_23_1", "3_7_3_1", "3_5_33_4"})
	assertCards(t, game.Table.Cards[1], []string{"2_6_2", "2_3_22_1", "2_23_2_4", "2_5_3_2"})
	assertCards(t, game.Table.Cards[0], []string{"1_4_3", "1_22_1", "1_22_0", "1_2_31_3"})

	// TODO: Make some moves.
}

func assertGameState(t *testing.T, game *Game, id string, state string, current string) {
	if game.ID != id {
		t.Errorf("bad ID: expected %v, got %v", id, game.ID)
	}
	if game.State != state {
		t.Errorf("bad state: expected %v, got %v", state, game.State)
	}
	if game.Current != current {
		t.Errorf("bad current player: expected %v, got %v", current, game.Current)
	}
}

func assertCoins(t *testing.T, game *Game, coins map[string]int) {
	if len(coins) != len(game.Table.Coins) {
		t.Errorf("bad table coins: expected %v, got %v", coins, game.Table.Coins)
		return
	}

	for k, v := range coins {
		if game.Table.Coins[k] != v {
			t.Errorf("bad table coins: expected %v, got %v", coins, game.Table.Coins)
			return
		}
	}
}

func assertNobles(t *testing.T, game *Game, nobles []string) {
	actual := []string{}
	for _, n := range game.Table.Nobles {
		actual = append(actual, n.ID)
	}
	if len(nobles) != len(actual) {
		t.Errorf("bad table nobles: expected %v, got %v", nobles, actual)
		return
	}
	for i, n := range actual {
		if n != nobles[i] {
			t.Errorf("bad table nobles: expected %v, got %v", nobles, actual)
			return
		}
	}
}

func assertCards(t *testing.T, cards []*Card, expected []string) {
	actual := []string{}
	for _, c := range cards {
		actual = append(actual, c.ID)
	}
	if len(expected) != len(actual) {
		t.Errorf("bad cards: expected %v, got %v", expected, actual)
		return
	}
	for i, a := range actual {
		if a != expected[i] {
			t.Errorf("bad cards: expected %v, got %v", expected, actual)
			return
		}
	}
}

func setup(url string) (*Impl, error) {
	if url == "" {
		cleanup(url)

		cmd := exec.Command("createdb", "splenda-test")
		cmd.Run()
		url = "postgres://localhost/splenda-test"
	}
	url += "?sslmode=disable"

	db := NewDB(url)
	if err := db.ApplySchema(); err != nil {
		return nil, err
	}

	auth, err := NewAuth(db, "aaa=")
	if err != nil {
		return nil, err
	}

	if err := auth.NewUser("user1", "abc"); err != nil {
		return nil, err
	}
	if err := auth.NewUser("user2", "def"); err != nil {
		return nil, err
	}

	impl := NewImplSeed(db, 1)
	return impl, nil
}

func cleanup(url string) {
	if url == "" {
		cmd := exec.Command("dropdb", "splenda-test")
		cmd.Run()
	}
}
