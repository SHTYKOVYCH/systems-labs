package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
)

func main() {
	cred := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      "root",
		Password:      "example",
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://db-notifications:27017").SetAuth(cred))
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
		"hello",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println("Error on opening queue")
		fmt.Println(err.Error())
		return
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		fmt.Println("Error on opening consume")
		fmt.Println(err.Error())
		return
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range msgs {
			fmt.Println(string(msg.Body))
		}
	}()

	wg.Wait()
}
