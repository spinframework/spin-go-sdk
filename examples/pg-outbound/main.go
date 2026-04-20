package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
	"github.com/spinframework/spin-go-sdk/v3/pg"
)

type Pet struct {
	ID        int64
	Name      string
	Prey      *string // nullable field must be a pointer
	IsFinicky bool
	Timestamp time.Time
}

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {

		// addr is the environment variable set in `spin.toml` that points to the
		// address of the postgres server.
		addr := os.Getenv("DB_URL")

		db := pg.Open(addr)
		defer db.Close()

		_, err := db.Query("INSERT INTO pets VALUES ($1, 'Maya', $2, $3, $4);", int32(4), "bananas", true, time.Now())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rows, err := db.Query("SELECT * FROM pets")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var pets []*Pet
		for rows.Next() {
			var pet Pet
			if err := rows.Scan(&pet.ID, &pet.Name, &pet.Prey, &pet.IsFinicky, &pet.Timestamp); err != nil {
				fmt.Println(err)
			}
			pets = append(pets, &pet)
		}
		json.NewEncoder(w).Encode(pets)
	})
}

func main() {}
