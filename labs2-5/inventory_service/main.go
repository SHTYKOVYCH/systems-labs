package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"net/http"
	"os"
)

type Product struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	NumOfItems int    `json:"numOfItems"`
	Price      int    `json:"price"`
}

func main() {
	db, err := sql.Open("postgres", "host='db-inventory' sslmode=disable port=5432 user=inventory dbname='inventory' password='inventory'")

	if err != nil {
		fmt.Println("Cannot connect to db")
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS items (\n    id serial PRIMARY KEY,\n    name VARCHAR NOT NULL,\n    numOfItems NUMERIC NOT NULL,\n    price int NOT NULL\n)")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get(os.Getenv("api_url")+"/items", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM items")

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error fetching data from database"))
			return
		}
		defer rows.Close()
		items := []Product{}
		var item Product
		for rows.Next() {
			err := rows.Scan(&item.Id, &item.Name, &item.NumOfItems, &item.Price)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Error fetching data from rows"))
				return
			}
			items = append(items, item)
		}

		res, err := json.Marshal(items)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Server error"))
			return
		}

		w.Write(res)
	})

	router.Post(os.Getenv("api_url")+"/items", func(w http.ResponseWriter, r *http.Request) {
		var bodyTraslated Product

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&bodyTraslated)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error on parsing body"))
			w.Write([]byte(err.Error()))
			return
		}

		_, err := db.Exec("INSERT INTO items (name, numOfItems, price) VALUES ($1, $2, $3)", bodyTraslated.Name, bodyTraslated.NumOfItems, bodyTraslated.Price)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error creating data from database\n"))
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte("ok"))
	})

	router.Patch(os.Getenv("api_url")+"/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		productId := chi.URLParam(r, "id")

		var bodyTraslated Product

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&bodyTraslated)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error on parsing body"))
			w.Write([]byte(err.Error()))
			return
		}

		_, err := db.Exec("UPDATE items SET name=$2, weight=$3, description=$4 WHERE id=$1", productId, bodyTraslated.Name, bodyTraslated.NumOfItems, bodyTraslated.Price)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error creating data from database\n"))
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte("ok"))
	})

	router.Delete(os.Getenv("api_url")+"/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		productId := chi.URLParam(r, "id")
		_, err = db.Exec("DELETE FROM items WHERE id=$1", productId)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error on deleting\n"))
			w.Write([]byte(err.Error()))
		}

		w.Write([]byte("ok"))
	})

	err = http.ListenAndServe(":8080", router)

	if err != nil {
		fmt.Println(err.Error())
	}
}
