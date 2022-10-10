package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

const webPort = "8080"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	// try to connect to rabbitmq
	rabbitConn, err := connct()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("We connected to rabbitMQ successfully!")
	defer rabbitConn.Close()
	// start listening for the messages
	log.Println("Listening for consuming RabbitMQ messages...")

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server

	err = srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}

func connct() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// don't continue until rabbit is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")

		if err != nil {
			fmt.Println("RabbitMQ is not ready ...")
		} else {
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off ...")
		time.Sleep(backOff)
		counts++
		continue
	}

	return connection, nil
}
