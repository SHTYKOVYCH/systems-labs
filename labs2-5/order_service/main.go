package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"slices"
	"time"
)

type Order struct {
	Id       primitive.ObjectID `bson:"_id"`
	Datetime time.Time          `bson:"datetime"`
}

type OrderItem struct {
	OrderId primitive.ObjectID `bson:"order_id"`
	ID      primitive.ObjectID `bson:"_id"`
	Name    string             `bson:"item_name"`
	Count   int                `bson:"count"`
	Price   int                `bson:"price"`
}

type OrderPostItem struct {
	ProductId int `json:"product_id"`
	Count     int `json:"count"`
	Price     int `json:"price"`
}

type PostBody struct {
	Items []OrderPostItem `json:"items"`
}

type Product struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	NumOfItems int    `json:"numOfItems"`
	Price      int    `json:"price"`
}

type Message struct {
	TypeOfMessage string    `json:"type"`
	Description   string    `json:"description"`
	Datetime      time.Time `json:"datetime"`
}

func main() {
	cred := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      "root",
		Password:      "example",
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://db-orders:27017").SetAuth(cred))
	if err != nil {
		fmt.Println("Cannot connect to db")
		fmt.Println(err.Error())
		return
	}

	defer client.Disconnect(context.TODO())

	rab, err := amqp.Dial("amqp://guest:guest@broker:5672/")

	if err != nil {
		fmt.Println("Error on connecting to rabbitmq")
		fmt.Println(err.Error())
		return
	}

	defer rab.Close()

	ch, err := rab.Channel()

	if err != nil {
		fmt.Println("Error on opening chanel")
		fmt.Println(err.Error())
		return
	}

	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		fmt.Println("Error on opening queue")
		fmt.Println(err.Error())
		return
	}

	orders := client.Database("orders").Collection("orders")
	products := client.Database("orders").Collection("products")

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/orders", func(w http.ResponseWriter, r *http.Request) {
		result, err := orders.Find(context.TODO(), bson.D{})

		if err == mongo.ErrNoDocuments {
			w.Write([]byte("[]"))
			return
		}

		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(500)
			w.Write([]byte("Error on reading from database"))
			return
		}

		defer result.Close(context.TODO())

		returning := []Order{}

		for result.Next(context.TODO()) {
			var res Order
			err := result.Decode(&res)

			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Error on decoding"))
				return
			}

			returning = append(returning, res)
		}

		returningStr, err := json.Marshal(returning)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error on marshal"))
			return
		}

		w.Write(returningStr)
	})

	router.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
		var bodyTranslated PostBody

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&bodyTranslated)

		resp, err := http.Get("http://inventory:8080/items")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error on fetching products"))
			return
		}
		var productsCatalog []Product
		pr_decoder := json.NewDecoder(resp.Body)
		err = pr_decoder.Decode(&productsCatalog)

		var resultItems []OrderItem
		orderId := primitive.NewObjectID()

		for _, el := range bodyTranslated.Items {
			index := slices.IndexFunc(productsCatalog, func(c Product) bool { return c.Id == el.ProductId })
			if index == -1 {
				w.WriteHeader(400)
				w.Write([]byte("Bad product id"))
				return
			}

			if productsCatalog[index].NumOfItems < el.Count {
				w.WriteHeader(400)
				w.Write([]byte("Too many items"))
				return
			}

			resultItems = append(resultItems, OrderItem{ID: primitive.NewObjectID(), OrderId: orderId, Name: productsCatalog[index].Name, Count: el.Count, Price: el.Price})
		}

		for _, el := range resultItems {

			_, err = products.InsertOne(context.TODO(), el)

			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Error on writing"))
				fmt.Println(err.Error())
				return
			}
		}

		_, err = orders.InsertOne(context.TODO(), Order{Datetime: time.Now(), Id: orderId})

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error on writing"))
			fmt.Println(err.Error())
			return
		}

		msg := Message{TypeOfMessage: "Create", Description: "Order created", Datetime: time.Now()}

		msgJSON, _ := json.Marshal(msg)
		err = ch.PublishWithContext(context.TODO(),
			"",
			queue.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(msgJSON),
			})

		if err != nil {
			fmt.Println("Error on publishing message")
			fmt.Println(err.Error())
			w.WriteHeader(500)
		}
	})

	router.Delete("/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
		_, err = orders.DeleteOne(context.TODO(), bson.D{{"_id", id}})
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error on delete"))
			fmt.Println(err.Error())
			return
		}
	})

	err = http.ListenAndServe(":8080", router)

	if err != nil {
		fmt.Println(err.Error())
	}
}
