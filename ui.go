package splenda

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"
)

// Root handles GET /, loading the homepage (authenticated or unauthenticated).
func (a *api) Root(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(405)
		return
	}

	_, err := a.authorize(req)
	if err != nil {
		renderFile("web/index.html", res)
	} else {
		renderFile("web/index-authenticated.html", res)
	}
}

// Asset handles requests for arbitrary public assets from the web folder.
func (a *api) Asset(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(405)
		return
	}
	file := path.Join("web", strings.TrimPrefix(req.URL.Path, "/assets/"))
	renderFile(file, res)
}

// Signup handles GET /signup, rendering the UI for signing up a new user.
func (a *api) Signup(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(405)
		return
	}
	renderFile("web/signup.html", res)
}

// Login handles POST /login, the browser-based log-in endpoint.
func (a *api) Login(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(405)
		return
	}

	if err := req.ParseForm(); err != nil {
		res.WriteHeader(400)
		return
	}

	id := req.Form.Get("id")
	pw := req.Form.Get("pw")

	sid, err := a.auth.Login(id, pw)
	if err != nil {
		renderFile("web/login-failed.html", res)
		return
	}

	// Redirect back to the homepage with a session cookie.
	http.SetCookie(res, &http.Cookie{
		Name:    "sid",
		Value:   sid,
		Path:    "/",
		Expires: time.Now().Add(13 * 24 * time.Hour),
	})
	res.Header().Set("Location", "/")
	res.WriteHeader(303)
}

// Logout handles POST /logout, the browser-based log-out endpoint.
func (a *api) Logout(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(405)
		return
	}

	// Redirect back to the homepage with NO session cookie.
	http.SetCookie(res, &http.Cookie{
		Name:    "sid",
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-1 * time.Hour),
	})
	res.Header().Set("Location", "/")
	res.WriteHeader(303)
}

// Game handles GET /games/<id>, rendering the game view.
func (a *api) Game(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(405)
		return
	}

	userID, err := a.authorize(req)
	if err != nil {
		res.WriteHeader(401)
		return
	}

	gameID := strings.TrimPrefix(req.URL.Path, "/games/")

	switch req.URL.Query().Get("v") {
	case "", "1":
		renderTemplate("web/game.html", gameID, userID, res)
	case "2":
		renderTemplate("web/game2.html", gameID, userID, res)
	default:
		res.WriteHeader(400)
	}
}

// RenderFile renders a non-template file.
func renderFile(file string, res http.ResponseWriter) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	renderBytes(file, bs, res)
}

// RenderTemplate renders a templated file, subbing in {{GAMEID}} and {{USERID}}.
func renderTemplate(file string, gameID string, userID string, res http.ResponseWriter) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error() + "\n"))
		return
	}

	bs = bytes.ReplaceAll(bs, []byte("{{GAMEID}}"), []byte(gameID))
	bs = bytes.ReplaceAll(bs, []byte("{{USERID}}"), []byte(userID))

	renderBytes(file, bs, res)
}

// RenderBytes renders the pre-loaded bytes of a file with the appropriate MIME type.
func renderBytes(file string, bs []byte, res http.ResponseWriter) {
	if strings.HasSuffix(file, ".html") {
		res.Header().Set("Content-Type", "text/html")
	} else if strings.HasSuffix(file, ".js") {
		res.Header().Set("Content-Type", "text/javascript")
	} else if strings.HasSuffix(file, ".css") {
		res.Header().Set("Content-Type", "text/css")
	}

	res.WriteHeader(200)
	res.Write(bs)
}
