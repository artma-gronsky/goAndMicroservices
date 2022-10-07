package main

import (
	"context"
	"log"
	"logger-service/data"
	"time"
)

type RPCServer struct {
}
type RPCPayload struct {
	Name string
	Data string
}

func (r *RPCServer) LogInfo(payload RPCPayload, resp *string) error {
	collection := client.Database("loggerService").Collection("logs")

	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:     payload.Name,
		Data:     payload.Data,
		CreateAt: time.Now().UTC(),
	})

	if err != nil {
		log.Println("error writing to mongo: ", err)
		return err
	}

	*resp = "Processed payload by RPC: " + payload.Name

	return nil
}
