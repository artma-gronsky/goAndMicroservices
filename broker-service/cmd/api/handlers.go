package main

import (
	"broker/event"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-http-utils/headers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"logger-service/logs"
	"net/http"
	"net/rpc"
	"time"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		{
			app.authenticate(w, requestPayload.Auth)
			return
		}
	case "log":
		{
			//app.log(w, requestPayload.Log)
			//app.logViaRabbitMQ(w, requestPayload.Log)
			app.logViaRPC(w, requestPayload.Log)
		}
	case "mail":
		{
			app.sendMail(w, requestPayload.Mail)
		}
	default:
		{
			app.errorJSON(w, errors.New("unsupported action was provided"), http.StatusBadRequest)
			return
		}
	}
}
func (app *Config) log(w http.ResponseWriter, a LogPayload) {
	jsonDataBytes, err := json.MarshalIndent(a, "", "\t")

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// todo: move to environment variable
	url := "http://logger-service/log"

	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonDataBytes))
	request.Header.Set(headers.ContentType, "application/json")

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling log service"))
		return
	}

	response.Body = http.MaxBytesReader(w, response.Body, int64(1024))

	decoder := json.NewDecoder(response.Body)

	var logResp jsonResponse
	err = decoder.Decode(&logResp)

	if err != nil || logResp.Error {
		app.errorJSON(w, errors.New("error getting response from log service"))
		return
	}

	app.writeJson(w, http.StatusAccepted, jsonResponse{
		Message: "Success",
		Error:   false,
		Data:    logResp.Data,
	})
}

func (app *Config) logViaRabbitMQ(w http.ResponseWriter, a LogPayload) {
	jsonData, _ := json.MarshalIndent(&a, "", "\t")

	emitter, err := event.NewEventEmitter(app.Rabbit)

	if err != nil {
		app.errorJSON(w, err)
	}

	err = emitter.Push(string(jsonData), "log.INFO")

	if err != nil {
		app.errorJSON(w, errors.New("error getting response from log service"))
		return
	}

	app.writeJson(w, http.StatusAccepted, jsonResponse{
		Message: "Success",
		Error:   false,
		Data:    "Will be logged via rabbitMQ",
	})
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logViaRPC(w http.ResponseWriter, a LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	rpcPayload := RPCPayload{
		Name: a.Name,
		Data: a.Data,
	}

	var result string

	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "The log message was logged by remote procedure call",
	}

	app.writeJson(w, http.StatusAccepted, payload)
}
func (app *Config) logViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, canel := context.WithTimeout(context.Background(), time.Second)
	defer canel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged via gRPC"

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// todo: move to environment variable
	url := "http://authenticate-service/authenticate"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService jsonResponse

	dec := json.NewDecoder(response.Body)

	err = dec.Decode(&jsonFromService)

	if err != nil {
		app.errorJSON(w, errors.New("problem with decoding response"), http.StatusInternalServerError)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.writeJson(w, http.StatusAccepted, payload)
}
func (app *Config) sendMail(w http.ResponseWriter, a MailPayload) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	url := "http://mail-service/send"

	log.Println(bytes.NewBuffer(jsonData))
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	request.Header.Set(headers.ContentType, "application/json")
	defer request.Body.Close()

	if err != nil {
		app.errorJSON(w, err)
	}

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
	}

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New(fmt.Sprintf("mail-service unswred with the status = %d", response.StatusCode)))
	}

	dec := json.NewDecoder(response.Body)

	var decoded jsonResponse

	err = dec.Decode(&decoded)

	if err != nil {
		app.errorJSON(w, err)
	}

	app.writeJson(w, http.StatusAccepted, decoded)
}
