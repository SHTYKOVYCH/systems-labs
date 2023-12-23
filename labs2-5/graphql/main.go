package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"sync"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

type Order struct {
	Id       primitive.ObjectID `bson:"_id"`
	Datetime time.Time          `bson:"datetime"`
}

type OrderPostItem struct {
	ProductId int `json:"product_id"`
	Count     int `json:"count"`
	Price     int `json:"price"`
}

type PostBody struct {
	Items []OrderPostItem `json:"items"`
}

type ProductItem struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Weight      int    `json:"weight"`
	Description string `json:"description"`
}

func returnAllProducts() ([]ProductItem, error) {
	conn, err := http.Get("http://products:8080/products")
	if err != nil {
		fmt.Println("Error in connecting to product_service, error: ", err)
		return nil, nil
	}
	var connBody []ProductItem
	decoder := json.NewDecoder(conn.Body)
	err = decoder.Decode(&connBody)
	if err != nil {
		fmt.Println("Error in reading data from product_service, error: ", err)
		return nil, nil
	}
	return connBody, nil
}

func returnAllOrders() ([]Order, error) {
	conn, err := http.Get("http://orders:8080/orders")
	if err != nil {
		fmt.Println("Error in getting orders answer, error:", err.Error())
		return nil, nil
	}
	var connBody []Order
	decoder := json.NewDecoder(conn.Body)
	err = decoder.Decode(&connBody)
	if err != nil {
		fmt.Println("Error in reading body order_server answer, error:", err.Error())
		return nil, nil
	}
	return connBody, nil
}

func sendToOrderService(products []OrderPostItem) (string, error) {
	jsonMess, err := json.Marshal(products)
	if err != nil {
		return "Error" + err.Error(), nil
	}
	dt := bytes.NewReader(jsonMess)
	_, err = http.Post("http://orders:8080/orders", "application/json", dt)
	if err != nil {
		return "Error on creating order" + err.Error(), nil
	}
	return "Ok", nil
}

func main() {
	productType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Product",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(ProductItem); ok {
						fmt.Println(product, ok)
						return product.Id, nil
					}
					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(ProductItem); ok {
						fmt.Println(product, ok)
						return product.Name, nil
					}
					return nil, nil
				},
			},
			"weight": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(ProductItem); ok {
						return product.Weight, nil
					}
					return nil, nil
				},
			},
			"description": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if product, ok := p.Source.(ProductItem); ok {
						fmt.Println(product, ok)
						return product.Description, nil
					}
					return nil, nil
				},
			},
		},
	})

	orderType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if orderProduct, ok := p.Source.(Order); ok {
						return orderProduct.Id, nil
					}
					return nil, nil
				},
			},
			"datetime": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if orderProduct, ok := p.Source.(Order); ok {
						return orderProduct.Datetime, nil
					}
					return nil, nil
				},
			},
		},
	})

	orderItemType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "OrderItem",
		Fields: graphql.InputObjectConfigFieldMap{
			"OrderId": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"name": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"count": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"price": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"products": &graphql.Field{
				Type: graphql.NewList(productType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return returnAllProducts()
				},
			},
			"order": &graphql.Field{
				Type: graphql.NewList(orderType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return returnAllOrders()
				},
			},
		},
	})
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createOrder": &graphql.Field{
				Type: graphql.String,
				Args: graphql.FieldConfigArgument{
					"items": &graphql.ArgumentConfig{
						Type: graphql.NewList(orderItemType),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					products := p.Source.(PostBody)
					return sendToOrderService(products.Items)
				},
			},
		},
	})
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
		Types:    []graphql.Type{orderItemType},
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   true,
		Playground: false,
	})

	http.Handle("/productsAndOrders", h)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = http.ListenAndServe(":8080", nil)
	}()

	wg.Wait()
}
