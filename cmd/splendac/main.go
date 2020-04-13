package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/fernomac/splenda"
)

type args struct {
	url  string
	sid  string
	args []string
}

var cmds = map[string]func(*args){
	"signup": func(a *args) {
		if len(a.args) < 2 {
			fmt.Println("usage: splendac signup <username> <password>")
			return
		}

		err := post(a.url+"/api/users", "", splenda.Login{
			ID:       a.args[0],
			Password: a.args[1],
		}, nil)
		if err != nil {
			panic(err)
		}
	},

	"login": func(a *args) {
		if len(a.args) < 2 {
			fmt.Println("usage: splendac login <username> <password>")
			return
		}

		sid := splenda.SID{}
		err := post(a.url+"/api/login", "", splenda.Login{
			ID:       a.args[0],
			Password: a.args[1],
		}, &sid)
		if err != nil {
			panic(err)
		}

		if err := ioutil.WriteFile(".sid", []byte(sid.SID), os.FileMode(0600)); err != nil {
			panic(err)
		}
	},

	"users": func(a *args) {
		users := splenda.UserList{}
		if err := get(a.url+"/api/users", a.sid, &users); err != nil {
			panic(err)
		}

		for _, user := range users.Users {
			fmt.Println(user)
		}
	},

	"games": func(a *args) {
		games := splenda.GameList{}
		if err := get(a.url+"/api/games", a.sid, &games); err != nil {
			panic(err)
		}

		for _, game := range games.Games {
			fmt.Println(game.ID)
			for _, player := range game.Players {
				fmt.Println("\t", player)
			}
		}
	},

	"newgame": func(a *args) {
		result := splenda.GameSummary{}
		err := post(a.url+"/api/games", a.sid, splenda.GameSummary{
			Players: a.args,
		}, &result)
		if err != nil {
			panic(err)
		}

		fmt.Println(result.ID)
	},

	"game": func(a *args) {
		if len(a.args) < 1 {
			fmt.Println("usage: splendac game <id>")
			return
		}
		result := splenda.Game{}

		err := get(a.url+"/api/games/"+a.args[0], a.sid, &result)
		if err != nil {
			panic(err)
		}

		fmt.Printf("id: %v\tts: %v\tstate: %v\tcurrent: %v\n",
			result.ID, result.TS, result.State, result.Current)

		fmt.Println()
		fmt.Printf("coins: %v\n", result.Table.Coins)

		fmt.Println("nobles:")
		for _, n := range result.Table.Nobles {
			fmt.Printf("  %v\t%v\n", n.Points, n.Cost)
		}

		fmt.Println("tier 3:")
		for _, c := range result.Table.Cards[2] {
			fmt.Printf("  %v\t%v\t%v\n", c.Color, c.Points, c.Cost)
		}

		fmt.Println("tier 2:")
		for _, c := range result.Table.Cards[1] {
			fmt.Printf("  %v\t%v\t%v\n", c.Color, c.Points, c.Cost)
		}

		fmt.Println("tier 1:")
		for _, c := range result.Table.Cards[0] {
			fmt.Printf("  %v\t%v\t%v\n", c.Color, c.Points, c.Cost)
		}

		fmt.Println()
		fmt.Println("players:")

		for _, p := range result.Players {
			fmt.Println(" ", p.ID, ":", p.Points)
			fmt.Printf("    coins: %v\n", p.Coins)

			fmt.Println("    nobles:")
			for _, n := range p.Nobles {
				fmt.Printf("      %v\t%v\n", n.Points, n.Cost)
			}

			m := map[string]int{}
			for color, cards := range p.Cards {
				m[color] = len(cards)
			}
			fmt.Println("    cards:", m)

			fmt.Println("    reserved:")
			for _, r := range p.Reserved {
				fmt.Printf("      %v\t%v\t%v\n", r.Color, r.Points, r.Cost)
			}
		}
	},

	"rmgame": func(a *args) {
		if len(a.args) < 1 {
			fmt.Println("usage: splendac take3 <id>")
			return
		}

		err := delete(a.url+"/api/games/"+a.args[0], a.sid)
		if err != nil {
			panic(err)
		}
	},

	"take3": func(a *args) {
		if len(a.args) < 1 {
			fmt.Println("usage: splendac take3 <id> <color> <color> <color>")
			return
		}

		ts := splenda.TS{}
		err := post(a.url+"/api/games/"+a.args[0]+"/take3", a.sid, splenda.Take3{
			Colors: a.args[1:],
		}, &ts)
		if err != nil {
			panic(err)
		}

		fmt.Println(ts.TS)
	},

	"take2": func(a *args) {
		if len(a.args) < 2 {
			fmt.Println("usage: splendac take2 <id> <color>")
			return
		}

		ts := splenda.TS{}
		err := post(a.url+"/api/games/"+a.args[0]+"/take2", a.sid, splenda.Take2{
			Color: a.args[1],
		}, &ts)
		if err != nil {
			panic(err)
		}

		fmt.Println(ts.TS)
	},

	"reserve": func(a *args) {
		if len(a.args) < 3 {
			fmt.Println("usage: splendac reserve <id> <tier> <index>")
		}

		tier, err := strconv.Atoi(a.args[1])
		if err != nil {
			panic(err)
		}
		index, err := strconv.Atoi(a.args[2])
		if err != nil {
			panic(err)
		}

		ts := splenda.TS{}
		err = post(a.url+"/api/games/"+a.args[0]+"/reserve", a.sid, splenda.Buy{
			Tier:  tier,
			Index: index,
		}, &ts)
		if err != nil {
			panic(err)
		}

		fmt.Println(ts.TS)
	},

	"buy": func(a *args) {
		if len(a.args) < 3 {
			fmt.Println("usage: splendac buy <id> <tier> <index>")
		}

		tier, err := strconv.Atoi(a.args[1])
		if err != nil {
			panic(err)
		}
		index, err := strconv.Atoi(a.args[2])
		if err != nil {
			panic(err)
		}

		ts := splenda.TS{}
		err = post(a.url+"/api/games/"+a.args[0]+"/buy", a.sid, splenda.Buy{
			Tier:  tier,
			Index: index,
		}, &ts)
		if err != nil {
			panic(err)
		}

		fmt.Println(ts.TS)
	},
}

func (a *args) call(cmd string) {
	f, ok := cmds[cmd]
	if !ok {
		panic(fmt.Sprintf("bad command: %v", cmd))
	}

	f(a)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: splendac <verb>")
		return
	}

	url := os.Getenv("BASE_URL")
	if url == "" {
		url = "http://localhost:8080"
	}

	sid, err := ioutil.ReadFile(".sid")
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	args := args{
		url:  url,
		sid:  string(sid),
		args: os.Args[2:],
	}
	args.call(os.Args[1])
}
