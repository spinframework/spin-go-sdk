package main

import (
	"encoding/json"
	"net/http"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
	"github.com/spinframework/spin-go-sdk/v3/sqlite"
)

type Pet struct {
	ID        int64
	Name      string
	Prey      *string // nullable field must be a pointer
	IsFinicky bool
}

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		db := sqlite.Open("default")
		defer db.Close()

		_, err := db.Query("REPLACE INTO pets (id, name, prey, is_finicky) VALUES (4, 'Maya', ?, false);", "bananas")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rows, err := db.Query("SELECT id, name, prey, is_finicky FROM pets")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var pets []*Pet
		for rows.Next() {
			var pet Pet
			if err := rows.Scan(&pet.ID, &pet.Name, &pet.Prey, &pet.IsFinicky); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			pets = append(pets, &pet)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pets)
	})
}

func main() {}
