package splenda

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type clock interface {
	Now() time.Time
}

type realclock struct{}

func (realclock) Now() time.Time {
	return time.Now()
}

// Auth implements Splenda's user authentication.
type Auth struct {
	db *DB

	cur   string
	keys  map[string][]byte
	clock clock
}

// NewAuth creates a new authenticator with a single key version.
// TODO: Support multiple key versions.
func NewAuth(db *DB, keystr string) (*Auth, error) {
	key, err := base64.StdEncoding.DecodeString(keystr)
	if err != nil {
		return nil, fmt.Errorf("invalid keystr: %v", err)
	}

	return &Auth{
		db:  db,
		cur: "1",
		keys: map[string][]byte{
			"1": key,
		},
		clock: realclock{},
	}, nil
}

// ListUsers lists all the currently registered users.
func (a *Auth) ListUsers() ([]string, error) {
	return a.db.ListUsers()
}

// NewUser creates a new game user.
func (a *Auth) NewUser(id, pw string) error {
	if id == "" {
		return errors.New("bad request: no id")
	}
	if pw == "" {
		return errors.New("bad request: no password")
	}

	return a.db.NewUser(id, a.hashPW(pw))
}

// Login logs in a user and returns a session ID.
func (a *Auth) Login(id, pw string) (string, error) {
	if id == "" {
		return "", errors.New("bad request: no id")
	}
	if pw == "" {
		return "", errors.New("bad request: no password")
	}

	hash, err := a.db.GetUserHash(id)
	if err != nil {
		return "", err
	}

	if !a.verifyPW(pw, hash) {
		return "", errors.New("bad request: invalid password")
	}

	return a.generateSID(id), nil
}

// Authorize verifies a session ID to confirm that it's real, returning the
// associated user ID if it is.
func (a *Auth) Authorize(sid string) (string, error) {
	id, ok := a.verifySID(sid)
	if !ok {
		return "", errors.New("invalid session id")
	}
	return id, nil
}

// HashPW hashes a password for storage in the database.
func (a *Auth) hashPW(pw string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(hash)
}

// VerifyPW verifies a password against the hash from the database.
func (a *Auth) verifyPW(pw, hash string) bool {
	hb, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword(hb, []byte(pw)) == nil
}

type sid struct {
	ID      string `json:"i"`
	Expires int64  `json:"e"`
}

// GenerateSID generates a session ID for the given user ID.
func (a *Auth) generateSID(id string) string {
	sid := sid{
		ID:      id,
		Expires: a.clock.Now().Add(14 * 24 * time.Hour).Unix(),
	}

	bs, err := json.Marshal(&sid)
	if err != nil {
		panic(err)
	}

	key := a.keys[a.cur]
	sig := hash(bs, key)

	return a.cur +
		"." +
		base64.StdEncoding.EncodeToString(bs) +
		"." +
		base64.StdEncoding.EncodeToString(sig)
}

// VerifySID verifies that a session ID is legit and returns the associated user ID.
func (a *Auth) verifySID(in string) (string, bool) {
	parts := strings.Split(in, ".")
	if len(parts) != 3 {
		// SID is malformed.
		return "", false
	}

	key, ok := a.keys[parts[0]]
	if !ok {
		// SID prefix contains an invalid key ID.
		return "", false
	}

	bs, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		// SID body is invalid base-64.
		return "", false
	}

	sig, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		// SID signature is invalid base-64.
		return "", false
	}

	asig := hash(bs, key)
	if subtle.ConstantTimeCompare(sig, asig) != 1 {
		// Signature mismatch.
		return "", false
	}

	sid := sid{}
	if err := json.Unmarshal(bs, &sid); err != nil {
		// Signature matches but body is invalid JSON; lolwut?
		return "", false
	}

	expires := time.Unix(sid.Expires, 0)
	if expires.Before(a.clock.Now()) {
		// SID has expired.
		return "", false
	}

	return sid.ID, true
}

func hash(bs, key []byte) []byte {
	hasher := hmac.New(sha256.New, key)
	if _, err := hasher.Write(bs); err != nil {
		panic(err)
	}
	return hasher.Sum(nil)
}
