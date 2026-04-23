package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
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

		if err := setupDB(db); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Testing Array parsing
		var x []int32
		if err := db.QueryRow(`SELECT ARRAY[200, 404]`).Scan(&x); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !slices.Equal(x, []int32{200, 404}) {
			http.Error(w, fmt.Sprintf("Slices aren't equal, got: %v", x), http.StatusInternalServerError)
			return
		}

		// Testing Range parsing
		var rangeInt32 pg.Int32Range
		if err := db.QueryRow(`SELECT int4range(10, 20)`).Scan(&rangeInt32); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if *rangeInt32.Lower != 10 {
			http.Error(w, fmt.Sprintf("Error parsing lower range, got: %v", *rangeInt32.Lower), http.StatusInternalServerError)
			return
		}
		if *rangeInt32.Upper != 20 {
			http.Error(w, fmt.Sprintf("Error parsing upper range, got: %v", *rangeInt32.Upper), http.StatusInternalServerError)
			return
		}

		_, err := db.Exec("INSERT INTO pets (id, name, prey, is_finicky, timestamp) VALUES ($1, 'Maya', $2, $3, $4);", int32(4), "bananas", true, time.Now())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rows, err := db.Query("SELECT * FROM pets")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var pets []*Pet
		for rows.Next() {
			var pet Pet
			if err := rows.Scan(&pet.ID, &pet.Name, &pet.Prey, &pet.IsFinicky, &pet.Timestamp); err != nil {
				fmt.Println(err)
			}
			pets = append(pets, &pet)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(pets); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

// temp function to make developing types easier. I'll delete this or the
// migration file before merging.
func setupDB(db *sql.DB) error {
	if _, err := db.Exec(`DROP TABLE pets`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE pets (
	  id INT PRIMARY KEY,
	  name VARCHAR(100) NOT NULL,
	  prey VARCHAR(100),
	  is_finicky BOOL NOT NULL,
	  timestamp TIMESTAMP
	)`); err != nil {
		return err
	}
	if _, err := db.Exec(`INSERT INTO pets VALUES (1, 'Splodge', NULL, false, '2026-04-20 12:30:00')`); err != nil {
		return err
	}
	if _, err := db.Exec(`INSERT INTO pets VALUES (2, 'Kiki', 'Cicadas', false, '2026-04-20 12:30:00')`); err != nil {
		return err
	}
	if _, err := db.Exec(`INSERT INTO pets VALUES (3, 'Slats', 'Temptations', true, '2026-04-20 12:30:00')`); err != nil {
		return err
	}

	return nil
}

func main() {}
