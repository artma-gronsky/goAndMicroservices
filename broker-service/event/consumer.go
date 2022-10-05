package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-http-utils/headers"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()

	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (c *Consumer) setup() error {
	channel, err := c.conn.Channel()

	if err != nil {
		return err
	}

	return declareExchange(channel)

}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	queue, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		err := ch.QueueBind(
			queue.Name,
			s,
			"logs_topic",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	message, err := ch.Consume(queue.Name, "", true, false, false, false, nil)

	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for d := range message {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message [Exchnage, Queue] [logs_topic, %s]", queue.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		// log whatever we got
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "auth":
		// authenticate

		// you can have as many cases as you want

	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func logEvent(entry Payload) error {
	jsonDataBytes, err := json.MarshalIndent(entry, "", "\t")

	if err != nil {
		return nil
	}

	// todo: move to environment variable
	url := "http://logger-service/log"

	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonDataBytes))
	request.Header.Set(headers.ContentType, "application/json")

	if err != nil {
		return err
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	decoder := json.NewDecoder(response.Body)

	type jsonResponse struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data    any    `json:"data,omitempty"`
	}

	var logResp jsonResponse
	err = decoder.Decode(&logResp)

	if err != nil || logResp.Error {
		return err
	}

	return nil
}
