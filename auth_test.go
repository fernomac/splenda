package splenda

import (
	"fmt"
	"testing"
	"time"
)

type mockclock struct {
	now time.Time
}

func (m *mockclock) Now() time.Time {
	return m.now
}

var (
	longLongAgo = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	nowish      = time.Date(2020, 3, 29, 22, 53, 12, 0, time.UTC)
	laterish    = time.Date(2020, 3, 31, 4, 20, 60, 0, time.UTC)
	theFuture   = time.Date(2020, 7, 1, 0, 0, 0, 0, time.UTC)
)

const (
	oldsid   = "1.eyJpIjoiYWJjIiwiZSI6LTYyMTY2MDA5NjAwfQ==.Wjr35mwUNU/9hRqitgZJvwo+2BfKEZF5ia76glcd0po="
	validsid = "1.eyJpIjoiZGF2aWQiLCJlIjoxNTg2NzMxOTkyfQ==.y3/vW1GdbVc76WImuT1R7+Y+MqCsMwJQnztBEzYxg/Q="
	badsig   = "1.eyJpIjoiZGF2aWQiLCJlIjoxNTg2NzMxOTkyfQ==.y3/vW1GdbVc76WImuT1R7+Y+MqCsMwJQnztBEzYxg/q="
)

func TestGenerateSID(t *testing.T) {
	clock := mockclock{}
	auth := &Auth{
		cur: "1",
		keys: map[string][]byte{
			"1": []byte{0},
		},
		clock: &clock,
	}

	test := func(id string, now time.Time, eval string) {
		clock.now = now
		t.Run(fmt.Sprintf("%v:%v", id, now.String()), func(t *testing.T) {
			val := auth.generateSID(id)
			if val != eval {
				t.Errorf("expected %v, got %v", eval, val)
			}
		})
	}

	test("abc", longLongAgo, oldsid)
	test("david", nowish, validsid)
}

func TestVerifySID(t *testing.T) {
	clock := mockclock{}
	auth := &Auth{
		cur: "1",
		keys: map[string][]byte{
			"1": []byte{0},
		},
		clock: &clock,
	}

	test := func(sid string, now time.Time, eid string, eok bool) {
		clock.now = now
		t.Run(sid, func(t *testing.T) {
			id, ok := auth.verifySID(sid)
			if ok != eok {
				t.Errorf("expected ok=%v, got %v", eok, ok)
			}
			if id != eid {
				t.Errorf("expected id=%v, got %v", eid, id)
			}
		})
	}

	test("bogus", nowish, "", false)
	test("bogus.aaa=.aaa=", nowish, "", false)
	test("1.bogus.aaa=", nowish, "", false)
	test("1.aaa=.aaa=", nowish, "", false)

	test(validsid, nowish, "david", true)
	test(badsig, nowish, "", false)

	test(oldsid, longLongAgo, "abc", true)
	test(oldsid, nowish, "", false)
}
