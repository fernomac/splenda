package splenda

import (
	"os/exec"
	"testing"
)

func TestTwoPlayers(t *testing.T) {
	defer cleanup()

	impl, err := setup()
	if err != nil {
		t.Fatal(err)
	}

	id, err := impl.NewGame([]string{"user1", "user2"})
	if err != nil {
		t.Fatal(err)
	}

	game, err := impl.GetGame(id, "user1", "0")
	if err != nil {
		t.Fatal(err)
	}
	assertGameState(t, game, id, play, "user1")
	assertCoins(t, game, map[string]int{red: 4, green: 4, blue: 4, black: 4, white: 4, wild: 5})
}

func assertGameState(t *testing.T, game *Game, id string, state string, current string) {
	if game.ID != id {
		t.Errorf("bad ID: expected %v, got %v", id, game.ID)
	}
	if game.State != play {
		t.Errorf("bad state: expected %v, got %v", play, game.State)
	}
	if game.Current != "user1" {
		t.Errorf("bad current player: expected %v, got %v", "user1", game.Current)
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

func setup() (*Impl, error) {
	cleanup()

	cmd := exec.Command("createdb", "splenda-test")
	cmd.Run()

	db := NewDB("postgres://localhost/splenda-test?sslmode=disable")
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

	return NewImpl(db), nil
}

func cleanup() {
	cmd := exec.Command("dropdb", "splenda-test")
	cmd.Run()
}
