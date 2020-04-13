package splenda

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// An API instance hosts the Splenda HTTP API.
type API interface {
	Serve(addr string) error
}

// NewAPI creates a new API.
func NewAPI(auth *Auth, impl *Impl) API {
	return &api{auth, impl}
}

type api struct {
	auth *Auth
	impl *Impl
}

// Serve serves the API on the given endpoint.
func (a *api) Serve(port string) error {
	http.HandleFunc("/", a.Root)
	http.HandleFunc("/assets/", a.Asset)
	http.HandleFunc("/signup", a.Signup)
	http.HandleFunc("/login", a.Login)
	http.HandleFunc("/logout", a.Logout)
	http.HandleFunc("/games/", a.Game)

	http.HandleFunc("/api/users", a.UsersAPI)
	http.HandleFunc("/api/login", a.LoginAPI)
	http.HandleFunc("/api/games", a.GamesAPI)
	http.HandleFunc("/api/games/", a.GameAPI)

	return http.ListenAndServe(port, nil)
}

// Authorize checks if the request contains a valid session id and, if so,
// extracts and returns the userID.
func (a *api) authorize(req *http.Request) (string, error) {
	sid, err := req.Cookie("sid")
	if err != nil {
		return "", err
	}
	return a.auth.Authorize(sid.Value)
}

// UsersAPI dispatches GET|POST /api/users to the appropriate handler.
func (a *api) UsersAPI(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		a.ListUsersAPI(res, req)

	case http.MethodPost:
		a.NewUserAPI(res, req)

	default:
		res.WriteHeader(405)
	}
}

// ListUsersAPI handles GET /api/users, listing registered users.
func (a *api) ListUsersAPI(res http.ResponseWriter, req *http.Request) {
	_, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	users, err := a.auth.ListUsers()
	if err != nil {
		// TODO: better error handling?
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	write(UserList{users}, res)
}

// NewUserAPI handles POST /api/users, registering a new user.
func (a *api) NewUserAPI(res http.ResponseWriter, req *http.Request) {
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

	res.WriteHeader(201)
}

// LoginAPI handles POST /api/login, programmatically logging in to the service.
func (a *api) LoginAPI(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
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

	write(SID{sid}, res)
}

// GamesAPI dispatches GET|POST /api/games to the right handler.
func (a *api) GamesAPI(res http.ResponseWriter, req *http.Request) {
	id, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	switch req.Method {
	case http.MethodGet:
		a.ListGamesAPI(id, res, req)

	case http.MethodPost:
		a.NewGameAPI(id, res, req)

	default:
		res.WriteHeader(405)
	}
}

// ListGamesAPI handles GET /api/games, listing the user's currently-running games.
func (a *api) ListGamesAPI(userID string, res http.ResponseWriter, req *http.Request) {
	games, err := a.impl.ListGames(userID)
	if err != nil {
		// TODO: better error handling?
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	write(GameList{games}, res)
}

// NewGameAPI handles POST /api/games, starting a new game.
func (a *api) NewGameAPI(userID string, res http.ResponseWriter, req *http.Request) {
	game := GameSummary{}
	if err := unmarshal(req.Body, &game); err != nil {
		res.WriteHeader(400)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	if game.ID != "" {
		res.WriteHeader(400)
		res.Write([]byte("cannot set id\n"))
		return
	}

	id, err := a.impl.NewGame(userID, game.Players)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	game.ID = id

	write(game, res)
}

// GameAPI dispatches GET|POST|DELETE /api/games/<id> to the right handler.
func (a *api) GameAPI(res http.ResponseWriter, req *http.Request) {
	userID, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/api/games/")

	switch req.Method {
	case http.MethodGet:
		a.GetGameAPI(userID, path, res, req)

	case http.MethodDelete:
		a.DeleteGameAPI(userID, path, res, req)

	case http.MethodPost:
		a.MoveAPI(userID, path, res, req)

	default:
		res.WriteHeader(405)
		return
	}
}

// GetGameAPI handles GET /api/games/<id>, getting the current state of a particular game.
func (a *api) GetGameAPI(userID string, path string, res http.ResponseWriter, req *http.Request) {
	gameID := path
	ts := req.URL.Query().Get("ts")

	game, err := a.impl.GetGame(gameID, userID, ts)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	write(game, res)
}

// DeleteGameAPI handles DELETE /api/games/<id>, deleting a game.
func (a *api) DeleteGameAPI(userID string, path string, res http.ResponseWriter, req *http.Request) {
	gameID := path

	if err := a.impl.DeleteGame(gameID, userID); err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	res.WriteHeader(204)
}

// MoveAPI handles POST /api/games/<id>/<move>, performing a move.
func (a *api) MoveAPI(userID string, path string, res http.ResponseWriter, req *http.Request) {
	idx := strings.IndexByte(path, '/')
	if idx == -1 {
		res.WriteHeader(405)
		return
	}

	gameID := path[:idx]
	trailer := path[idx+1:]

	switch trailer {
	case "take3":
		a.Take3API(gameID, userID, res, req)

	case "take2":
		a.Take2API(gameID, userID, res, req)

	case "reserve":
		a.ReserveAPI(gameID, userID, res, req)

	case "buy":
		a.BuyAPI(gameID, userID, res, req)

	default:
		res.WriteHeader(404)
	}
}

// Take3API handles POST /api/games/<id>/take3, taking three coins from the table.
func (a *api) Take3API(gameID string, userID string, res http.ResponseWriter, req *http.Request) {
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

	write(TS{ts}, res)
}

// Take2API handles POST /games/<id>/take2, taking two coins from the table.
func (a *api) Take2API(gameID string, userID string, res http.ResponseWriter, req *http.Request) {
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

	write(TS{ts}, res)
}

// ReserveAPI handles POST /games/<id>/reserve, reserving a card.
func (a *api) ReserveAPI(gameID string, userID string, res http.ResponseWriter, req *http.Request) {
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

	write(TS{ts}, res)
}

// BuyAPI handles POST /games/<id>/buy, buying a card.
func (a *api) BuyAPI(gameID string, userID string, res http.ResponseWriter, req *http.Request) {
	move := Buy{}
	if err := unmarshal(req.Body, &move); err != nil {
		res.WriteHeader(400)
		return
	}

	ts, err := a.impl.Buy(gameID, userID, move.Tier, move.Index)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	write(TS{ts}, res)
}

func write(d interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	marshal(d, w)
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
