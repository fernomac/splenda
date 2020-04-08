package splenda

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// API implements the Splenda HTTP API.
type API struct {
	auth *Auth
	impl *Impl
}

// NewAPI creates a new API.
func NewAPI(auth *Auth, impl *Impl) *API {
	return &API{auth, impl}
}

// Serve serves the API on the given endpoint.
func (a *API) Serve(port string) error {
	http.HandleFunc("/", a.GetRoot)
	http.HandleFunc("/api/users", a.Users)
	http.HandleFunc("/api/login", a.Login)

	http.HandleFunc("/api/games", a.Games)
	http.HandleFunc("/api/games/", a.Game)

	return http.ListenAndServe(port, nil)
}

func (a *API) authorize(req *http.Request) (string, error) {
	sid, err := req.Cookie("sid")
	if err != nil {
		return "", err
	}
	return a.auth.Authorize(sid.Value)
}

// GetRoot handles GET / -- load the html, javascript, etc.
func (a *API) GetRoot(res http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		res.WriteHeader(405)
		return
	}

	id, err := a.authorize(req)
	if err != nil {
		a.getAnonymousRoot(res, req)
		return
	}

	a.getAuthorizedRoot(id, res, req)
}

func (a *API) getAnonymousRoot(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Hello, anonymous\n"))
}

func (a *API) getAuthorizedRoot(id string, res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(fmt.Sprintf("Hello, %v\n", id)))
}

// Users dispatches GET|POST /api/users.
func (a *API) Users(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		a.listUsers(res, req)

	case "POST":
		a.newUser(res, req)

	default:
		res.WriteHeader(405)
	}
}

// GET /api/users -- List registered users.
func (a *API) listUsers(res http.ResponseWriter, req *http.Request) {
	_, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	ids, err := a.auth.ListUsers()
	if err != nil {
		// TODO: better error handling?
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	marshal(UserList{Users: ids}, res)
}

// POST /api/users -- Register a new user.
func (a *API) newUser(res http.ResponseWriter, req *http.Request) {
	login := Login{}
	if err := unmarshal(req.Body, &login); err != nil {
		// TODO: better error handling?
		res.WriteHeader(400)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	err := a.auth.NewUser(login.ID, login.Password)
	if err != nil {
		// TODO: better error handling?
		res.WriteHeader(400)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	loc := fmt.Sprintf("/api/users/%v", url.PathEscape(login.ID))
	res.Header().Set("Location", loc)
	res.Header().Set("Content-Length", "0")
	res.WriteHeader(201)
}

// Login handles POST /api/login -- log in to the service.
func (a *API) Login(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		res.WriteHeader(405)
		return
	}

	login := Login{}
	if err := unmarshal(req.Body, &login); err != nil {
		res.WriteHeader(400)
		return
	}

	sid, err := a.auth.Login(login.ID, login.Password)
	if err != nil {
		// TODO: better error handling?
		res.WriteHeader(400)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	// Return a 200 with a session cookie.
	http.SetCookie(res, &http.Cookie{
		Name:    "sid",
		Value:   sid,
		Path:    "/",
		Expires: time.Now().Add(13 * 24 * time.Hour),
	})
	res.Header().Set("Content-Length", "0")
	res.WriteHeader(200)
}

// Games handles GET|POST /api/games
func (a *API) Games(res http.ResponseWriter, req *http.Request) {
	id, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	switch req.Method {
	case "GET":
		a.listGames(id, res, req)

	case "POST":
		a.newGame(id, res, req)

	default:
		res.WriteHeader(405)
	}
}

// GET /api/games -- list currently-running games.
func (a *API) listGames(id string, res http.ResponseWriter, req *http.Request) {
	games, err := a.impl.ListGames(id)
	if err != nil {
		// TODO: better error handling?
		res.WriteHeader(400)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	dto := GameList{}
	for id, players := range games {
		dto.Games = append(dto.Games, &GameSummary{
			ID:      id,
			Players: players,
		})
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	marshal(dto, res)
}

// POST /games -- start a new game.
func (a *API) newGame(id string, res http.ResponseWriter, req *http.Request) {
	game := GameSummary{}
	if err := unmarshal(req.Body, &game); err != nil {
		res.WriteHeader(400)
		return
	}

	if game.ID != "" {
		res.WriteHeader(400)
		res.Write([]byte("cannot set id\n"))
		return
	}

	id, err := a.impl.NewGame(game.Players)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	game.ID = id

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	marshal(game, res)
}

// Game handles GET|POST|DELETE /api/games/...
func (a *API) Game(res http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/games/")

	if req.Method == http.MethodGet {
		a.getGame(path, res, req)
		return
	}
	if req.Method == http.MethodDelete {
		a.deleteGame(path, res, req)
		return
	}
	if req.Method != http.MethodPost {
		res.WriteHeader(405)
		return
	}

	idx := strings.IndexByte(path, '/')
	if idx == -1 {
		res.WriteHeader(405)
		return
	}

	gameID := path[:idx]
	trailer := path[idx+1:]

	switch trailer {
	case "take3":
		a.take3(gameID, res, req)

	case "take2":
		a.take2(gameID, res, req)

	case "reserve":
		a.reserve(gameID, res, req)

	default:
		res.WriteHeader(404)
	}
}

// GetGame handles GET /api/games/<id>?ts=<ts> -- get the current state of a particular game.
func (a *API) getGame(gameID string, res http.ResponseWriter, req *http.Request) {
	userID, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	ts := req.URL.Query().Get("ts")

	game, err := a.impl.GetGame(gameID, userID, ts)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	marshal(game, res)
}

// DELETE /api/games/<id> -- delete a game.
func (a *API) deleteGame(gameID string, res http.ResponseWriter, req *http.Request) {
	userID, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	if err := a.impl.DeleteGame(gameID, userID); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
}

// Take3 handles POST /api/games/<id>/take3 -- take three coins from the table.
func (a *API) take3(gameID string, res http.ResponseWriter, req *http.Request) {
	userID, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	move := Take3{}
	if err := unmarshal(req.Body, &move); err != nil {
		res.WriteHeader(400)
		return
	}

	ts, err := a.impl.Take3(gameID, userID, move.Colors)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	marshal(TS{ts}, res)
}

// Take2 handles POST /games/<id>/take2 -- take two coins from the table.
func (a *API) take2(gameID string, res http.ResponseWriter, req *http.Request) {
	userID, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	move := Take2{}
	if err := unmarshal(req.Body, &move); err != nil {
		res.WriteHeader(400)
		return
	}

	ts, err := a.impl.Take2(gameID, userID, move.Color)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	marshal(TS{ts}, res)
}

// Reserve handles POST /games/<id>/reserve -- reserve a card.
// {"level":1, "index": 2}
func (a *API) reserve(gameID string, res http.ResponseWriter, req *http.Request) {
	userID, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	move := Buy{}
	if err := unmarshal(req.Body, &move); err != nil {
		res.WriteHeader(400)
		return
	}

	ts, err := a.impl.Reserve(gameID, userID, move.Tier, move.Index)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	marshal(TS{ts}, res)
}

// Buy handles POST /games/<id>/buy -- buy a card.
func (a *API) Buy(res http.ResponseWriter, req *http.Request) {
	// TODO: verify login cookie.

	move := Buy{}
	if err := unmarshal(req.Body, &move); err != nil {
		res.WriteHeader(400)
		return
	}

	ts := TS{}
	// TODO: Query and update DB.

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	marshal(ts, res)
}

func marshal(src interface{}, dst io.Writer) {
	bs, err := json.Marshal(src)
	if err != nil {
		panic(err)
	}
	dst.Write(bs)
}

func unmarshal(src io.Reader, dst interface{}) error {
	bs, err := ioutil.ReadAll(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(bs, dst)
}
