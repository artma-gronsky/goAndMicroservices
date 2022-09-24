package data

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var client *mongo.Client

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        string    `bson:"_id,omitempty"json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreateAt  time.Time `bson:"create_at" json:"createAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

func New(mongo *mongo.Client) Models {
	client = mongo

	return Models{
		LogEntry: LogEntry{},
	}
}

func GetLoggerServiceLogCollection() *mongo.Collection {
	collection := client.Database("loggerService").Collection("logs")
	return collection
}

func (l *LogEntry) Insert(entry LogEntry) error {
	_, err := GetLoggerServiceLogCollection().InsertOne(context.TODO(), LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreateAt:  time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		log.Println("Error inserting on the logs:", err.Error())
		return err
	}

	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := GetLoggerServiceLogCollection()

	opts := options.Find()
	opts.SetSort(bson.D{{"created_at", -1}})

	cursor, err := collection.Find(ctx, bson.D{}, opts)

	if err != nil {
		log.Println("Finding all docs error: ", err.Error())
		return nil, err
	}

	defer cursor.Close(ctx)

	var logs []*LogEntry

	for cursor.Next(ctx) {
		var item LogEntry

		err := cursor.Decode(&item)

		if err != nil {
			log.Println("Error decoding log into slice: ", err.Error())
		} else {
			logs = append(logs, &item)
		}
	}

	return logs, nil
}

func (l *LogEntry) GetById(idHex string) (*LogEntry, error) {
	collection := GetLoggerServiceLogCollection()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(idHex)

	if err != nil {
		return nil, err
	}

	result := collection.FindOne(ctx, bson.M{"_id": objectId})

	var logEntry LogEntry

	if err := result.Err(); err != nil {
		return nil, err
	}

	err = result.Decode(logEntry)

	if err != nil {
		return nil, err
	}

	return &logEntry, nil
}

func (l *LogEntry) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5&time.Second)
	defer cancel()

	collection := GetLoggerServiceLogCollection()

	if err := collection.Drop(ctx); err != nil {
		return err
	}

	return nil
}

func (l *LogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5&time.Second)
	defer cancel()

	collection := GetLoggerServiceLogCollection()

	docId, err := primitive.ObjectIDFromHex(l.ID)

	if err != nil {
		return nil, err
	}

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": docId},
		bson.D{
			{"$set", bson.D{
				{"name", l.Name},
				{"data", l.Data},
				{"updated_at", time.Now().UTC()},
			}},
		})

	if err != nil {
		return nil, err
	}

	return result, nil

}
