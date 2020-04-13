package splenda

import (
	"sort"
	"strings"
)

const (
	white = "white"
	black = "black"
	green = "green"
	blue  = "blue"
	red   = "red"
	wild  = "wild"
)

func isNormalColor(color string) bool {
	for _, c := range []string{white, black, green, blue, red} {
		if c == color {
			return true
		}
	}
	return false
}

const (
	play      = "play"
	picknoble = "picknoble"
	losecoin  = "losecoin"
	gameover  = "gameover"
)

type cost map[string]int

type noble struct {
	id   string
	cost cost
}

func nobleFromID(id string) (noble, bool) {
	n, ok := nobles[id]
	return n, ok
}

type rng interface {
	Intn(n int) int
}

func shuffle(set []string, rng rng) []string {
	return pick(set, len(set), rng)
}

func pick(set []string, n int, rng rng) []string {
	ret := []string{}
	for i := 0; i < n; i++ {
		j := rng.Intn(len(set))
		ret = append(ret, set[j])
		set = append(set[:j], set[j+1:]...)
	}
	return ret
}

func pickNobles(n int, rng rng) []string {
	deck := []string{}
	for id := range nobles {
		deck = append(deck, id)
	}
	sort.Strings(deck) // To make things deterministic for tests.

	return pick(deck, n, rng)
}

type card struct {
	id     string
	color  string
	points int
	cost   cost
}

func cardFromID(id string) (card, bool) {
	if strings.HasPrefix(id, "1_") {
		c, ok := tier1[id]
		return c, ok
	}
	if strings.HasPrefix(id, "2_") {
		c, ok := tier2[id]
		return c, ok
	}
	if strings.HasPrefix(id, "3_") {
		c, ok := tier3[id]
		return c, ok
	}
	return card{}, false
}

func shuffleCards(cs map[string]card, rng rng) []string {
	deck := []string{}
	for id := range cs {
		deck = append(deck, id)
	}
	sort.Strings(deck) // To make things deterministic for tests.

	return pick(deck, len(deck), rng)
}

func noblemap(ns []noble) map[string]noble {
	ret := map[string]noble{}
	for i := range ns {
		ret[ns[i].id] = ns[i]
	}
	return ret
}

var nobles = noblemap([]noble{
	{id: "mary_stuart", cost: cost{red: 4, green: 4}},
	{id: "charles_v", cost: cost{black: 3, red: 3, white: 3}},
	{id: "macchiavelli", cost: cost{blue: 4, white: 4}},
	{id: "isabelle_of_castille", cost: cost{black: 4, white: 4}},
	{id: "suleiman_i", cost: cost{blue: 4, green: 4}},
	{id: "catherine_of_medici", cost: cost{green: 3, blue: 3, red: 3}},
	{id: "anne_of_brittany", cost: cost{green: 3, blue: 3, white: 3}},
	{id: "henry_viii", cost: cost{black: 4, red: 4}},
	{id: "elisabeth_of_austria", cost: cost{black: 3, blue: 3, white: 3}},
	{id: "francis_i", cost: cost{black: 3, red: 3, green: 3}},
})

func cardmap(cs []card) map[string]card {
	ret := map[string]card{}
	for i := range cs {
		ret[cs[i].id] = cs[i]
	}
	return ret
}

var tier1 = cardmap([]card{
	{id: "1_4_0", color: white, points: 1, cost: cost{green: 4}},
	{id: "1_4_1", color: green, points: 1, cost: cost{black: 4}},
	{id: "1_4_2", color: black, points: 1, cost: cost{blue: 4}},
	{id: "1_4_3", color: blue, points: 1, cost: cost{red: 4}},
	{id: "1_4_4", color: red, points: 1, cost: cost{white: 4}},

	{id: "1_3_0", color: white, cost: cost{blue: 3}},
	{id: "1_3_1", color: green, cost: cost{red: 3}},
	{id: "1_3_2", color: black, cost: cost{green: 3}},
	{id: "1_3_3", color: blue, cost: cost{black: 3}},
	{id: "1_3_4", color: red, cost: cost{white: 3}},

	{id: "1_2_1_0", color: white, cost: cost{red: 2, black: 1}},
	{id: "1_2_1_1", color: green, cost: cost{white: 2, blue: 1}},
	{id: "1_2_1_2", color: black, cost: cost{green: 2, red: 1}},
	{id: "1_2_1_3", color: blue, cost: cost{black: 2, white: 1}},
	{id: "1_2_1_4", color: red, cost: cost{blue: 2, green: 1}},

	{id: "1_22_0", color: white, cost: cost{blue: 2, black: 2}},
	{id: "1_22_1", color: green, cost: cost{blue: 2, red: 2}},
	{id: "1_22_2", color: black, cost: cost{white: 2, green: 2}},
	{id: "1_22_3", color: blue, cost: cost{green: 2, black: 2}},
	{id: "1_22_4", color: red, cost: cost{white: 2, red: 2}},

	{id: "1_41_0", color: white, cost: cost{blue: 1, green: 1, red: 1, black: 1}},
	{id: "1_41_1", color: green, cost: cost{blue: 1, white: 1, red: 1, black: 1}},
	{id: "1_41_2", color: black, cost: cost{blue: 1, green: 1, red: 1, white: 1}},
	{id: "1_41_3", color: blue, cost: cost{white: 1, green: 1, red: 1, black: 1}},
	{id: "1_41_4", color: red, cost: cost{blue: 1, green: 1, white: 1, black: 1}},

	{id: "1_3_21_0", color: white, cost: cost{white: 3, blue: 1, black: 1}},
	{id: "1_3_21_1", color: green, cost: cost{blue: 3, white: 1, green: 1}},
	{id: "1_3_21_2", color: black, cost: cost{red: 3, green: 1, black: 1}},
	{id: "1_3_21_3", color: blue, cost: cost{green: 3, red: 1, blue: 1}},
	{id: "1_3_21_4", color: red, cost: cost{black: 3, red: 1, white: 1}},

	{id: "1_22_1_0", color: white, cost: cost{blue: 2, green: 2, black: 1}},
	{id: "1_22_1_1", color: green, cost: cost{black: 2, red: 2, blue: 1}},
	{id: "1_22_1_2", color: black, cost: cost{white: 2, blue: 2, red: 1}},
	{id: "1_22_1_3", color: blue, cost: cost{red: 2, green: 2, white: 1}},
	{id: "1_22_1_4", color: red, cost: cost{white: 2, black: 2, green: 1}},

	{id: "1_2_31_0", color: white, cost: cost{green: 2, blue: 1, red: 1, black: 1}},
	{id: "1_2_31_1", color: green, cost: cost{black: 2, blue: 1, red: 1, white: 1}},
	{id: "1_2_31_2", color: black, cost: cost{blue: 2, white: 1, red: 1, green: 1}},
	{id: "1_2_31_3", color: blue, cost: cost{red: 2, white: 1, green: 1, black: 1}},
	{id: "1_2_31_4", color: red, cost: cost{white: 2, blue: 1, green: 1, black: 1}},
})

var tier2 = cardmap([]card{
	{id: "2_6_0", color: white, points: 3, cost: cost{white: 6}},
	{id: "2_6_1", color: green, points: 3, cost: cost{green: 6}},
	{id: "2_6_2", color: black, points: 3, cost: cost{black: 6}},
	{id: "2_6_3", color: blue, points: 3, cost: cost{blue: 6}},
	{id: "2_6_4", color: red, points: 3, cost: cost{red: 6}},

	{id: "2_5_0", color: white, points: 2, cost: cost{red: 5}},
	{id: "2_5_1", color: green, points: 2, cost: cost{green: 5}},
	{id: "2_5_2", color: black, points: 2, cost: cost{white: 5}},
	{id: "2_5_3", color: blue, points: 2, cost: cost{blue: 5}},
	{id: "2_5_4", color: red, points: 2, cost: cost{black: 6}},

	{id: "2_5_3_0", color: white, points: 2, cost: cost{red: 5, black: 3}},
	{id: "2_5_3_1", color: green, points: 2, cost: cost{blue: 5, green: 3}},
	{id: "2_5_3_2", color: black, points: 2, cost: cost{green: 5, red: 3}},
	{id: "2_5_3_3", color: blue, points: 2, cost: cost{white: 5, blue: 3}},
	{id: "2_5_3_4", color: red, points: 2, cost: cost{black: 5, white: 3}},

	{id: "2_4_2_1_0", color: white, points: 2, cost: cost{red: 4, black: 2, green: 1}},
	{id: "2_4_2_1_1", color: green, points: 2, cost: cost{white: 4, blue: 2, black: 1}},
	{id: "2_4_2_1_2", color: black, points: 2, cost: cost{green: 4, red: 2, blue: 1}},
	{id: "2_4_2_1_3", color: blue, points: 2, cost: cost{black: 4, white: 2, red: 1}},
	{id: "2_4_2_1_4", color: red, points: 2, cost: cost{blue: 4, green: 2, white: 1}},

	{id: "2_3_22_0", color: white, points: 1, cost: cost{green: 3, red: 2, black: 2}},
	{id: "2_3_22_1", color: green, points: 1, cost: cost{blue: 3, white: 2, black: 2}},
	{id: "2_3_22_2", color: black, points: 1, cost: cost{white: 3, blue: 2, green: 2}},
	{id: "2_3_22_3", color: blue, points: 1, cost: cost{red: 3, blue: 2, green: 2}},
	{id: "2_3_22_4", color: red, points: 1, cost: cost{black: 3, red: 2, white: 2}},

	{id: "2_23_2_0", color: white, points: 1, cost: cost{blue: 3, red: 3, white: 2}},
	{id: "2_23_2_1", color: green, points: 1, cost: cost{red: 3, white: 3, green: 2}},
	{id: "2_23_2_2", color: black, points: 1, cost: cost{white: 3, green: 3, black: 2}},
	{id: "2_23_2_3", color: blue, points: 1, cost: cost{green: 3, black: 3, blue: 2}},
	{id: "2_23_2_4", color: red, points: 1, cost: cost{blue: 3, black: 3, red: 2}},
})

var tier3 = cardmap([]card{
	{id: "3_7_3_0", color: white, points: 5, cost: cost{black: 7, white: 3}},
	{id: "3_7_3_1", color: green, points: 5, cost: cost{blue: 7, green: 3}},
	{id: "3_7_3_2", color: black, points: 5, cost: cost{red: 7, black: 3}},
	{id: "3_7_3_3", color: blue, points: 5, cost: cost{white: 7, blue: 3}},
	{id: "3_7_3_4", color: red, points: 5, cost: cost{green: 7, red: 3}},

	{id: "3_7_0", color: white, points: 4, cost: cost{black: 7}},
	{id: "3_7_1", color: green, points: 4, cost: cost{blue: 7}},
	{id: "3_7_2", color: black, points: 4, cost: cost{red: 7}},
	{id: "3_7_3", color: blue, points: 4, cost: cost{white: 7}},
	{id: "3_7_4", color: red, points: 4, cost: cost{green: 7}},

	{id: "3_6_23_0", color: white, points: 4, cost: cost{black: 6, white: 3, red: 3}},
	{id: "3_6_23_1", color: green, points: 4, cost: cost{blue: 6, green: 3, white: 3}},
	{id: "3_6_23_2", color: black, points: 4, cost: cost{red: 6, black: 3, green: 3}},
	{id: "3_6_23_3", color: blue, points: 4, cost: cost{white: 6, blue: 3, black: 3}},
	{id: "3_6_23_4", color: red, points: 4, cost: cost{green: 6, blue: 3, red: 3}},

	{id: "3_5_33_0", color: white, points: 3, cost: cost{red: 5, blue: 3, green: 3, black: 3}},
	{id: "3_5_33_1", color: green, points: 3, cost: cost{white: 5, blue: 3, red: 3, black: 3}},
	{id: "3_5_33_2", color: black, points: 3, cost: cost{green: 5, white: 3, blue: 3, red: 3}},
	{id: "3_5_33_3", color: blue, points: 3, cost: cost{black: 5, white: 3, green: 3, red: 3}},
	{id: "3_5_33_4", color: red, points: 3, cost: cost{blue: 5, white: 3, green: 3, black: 3}},
})
