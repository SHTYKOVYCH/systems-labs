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
	Code        string `json:"code"`
	Name        string `json:"name"`
	Weight      int    `json:"weight"`
	Description string `json:"description"`
}

func main() {
	db, err := sql.Open("postgres", "host='db-products' sslmode=disable port=5432 user=products dbname='products' password='products'")

	if err != nil {
		fmt.Println("Cannot connect to db")
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS products (\n    code VARCHAR PRIMARY KEY,\n    name VARCHAR NOT NULL,\n    weight NUMERIC NOT NULL,\n    description VARCHAR NOT NULL\n)")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get(os.Getenv("api_url")+"/products", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT code FROM products")

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error fetching data from database"))
			return
		}
		defer rows.Close()
		products := []string{}
		var code string
		for rows.Next() {
			err := rows.Scan(&code)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Error fetching data from rows"))
				return
			}
			products = append(products, code)
		}

		res, err := json.Marshal(products)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Server error"))
			return
		}

		w.Write(res)
	})

	router.Get(os.Getenv("api_url")+"/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		var code string
		err := db.QueryRow("SELECT code FROM products WHERE code=$1", chi.URLParam(r, "id")).Scan(&code)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error fetching data from database"))
			return
		}

		res, err := json.Marshal([]string{code})

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Server error"))
			return
		}

		w.Write(res)
	})

	router.Post(os.Getenv("api_url")+"/products", func(w http.ResponseWriter, r *http.Request) {
		var bodyTraslated Product

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&bodyTraslated)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error on parsing body"))
			w.Write([]byte(err.Error()))
			return
		}

		_, err := db.Exec("INSERT INTO products VALUES ($1, $2, $3, $4)", bodyTraslated.Code, bodyTraslated.Name, bodyTraslated.Weight, bodyTraslated.Description)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error creating data from database\n"))
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte("ok"))
	})

	router.Patch(os.Getenv("api_url")+"/products/{id}", func(w http.ResponseWriter, r *http.Request) {
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

		_, err := db.Exec("UPDATE products SET code=$1, name=$2, weight=$3, description=$4 WHERE code=$5", bodyTraslated.Code, bodyTraslated.Name, bodyTraslated.Weight, bodyTraslated.Description, productId)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error creating data from database\n"))
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte("ok"))
	})

	router.Delete(os.Getenv("api_url")+"/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		productId := chi.URLParam(r, "id")
		_, err = db.Exec("DELETE FROM products WHERE code=$1", productId)

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

	fmt.Println("Hi")
}
