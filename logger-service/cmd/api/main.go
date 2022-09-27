package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"logger-service/data"
	"net/http"
	"time"
)

const (
	webPort  = 80
	rpcPort  = 5001
	mongoUrl = "mongodb://mongo:27017"
	gRpcPort = 50001
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	//connect to mongo
	mongoClient, err := connectToMongo()

	if err != nil {
		log.Panic(err)
		return
	}
	client = mongoClient
	log.Println("Service successfully connected to mongo-db: ", mongoUrl)

	//create context
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// start web server
	//go app.server()
	log.Println("Server will be launched now on port:", webPort)
	app.server()
}

func (app *Config) server() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Panic("Can't start logging service: " + err.Error())
	}
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoUrl)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	c, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Println("Error connection: ", err)
		return nil, err
	}

	return c, nil
}
