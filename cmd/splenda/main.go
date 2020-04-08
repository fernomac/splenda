package main

import (
	"os"

	"github.com/fernomac/splenda"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://localhost/splenda?sslmode=disable"
	}

	db := splenda.NewDB(url)

	if len(os.Args) > 1 && os.Args[1] == "--schema" {
		if err := db.ApplySchema(); err != nil {
			panic(err)
		}
		return
	}

	keystr := os.Getenv("SID_KEY_1")
	if keystr == "" {
		keystr = "aaa="
	}

	auth, err := splenda.NewAuth(db, keystr)
	if err != nil {
		panic(err)
	}

	impl := splenda.NewImpl(db)
	api := splenda.NewAPI(auth, impl)

	port := os.Getenv("PORT")
	if port == "" {
		port = "localhost:8080"
	} else {
		port = ":" + port
	}

	api.Serve(port)
}
