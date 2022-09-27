package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-http-utils/headers"
	"log"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJson(w, err, http.StatusBadRequest)
		return
	}

	// validate the user against database

	user, err := app.Models.User.GetByEmail(requestPayload.Email)

	if err != nil {
		app.errorJson(w, errors.New("Invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)

	if err != nil || !valid {
		app.errorJson(w, errors.New("Invalid credentials"), http.StatusBadRequest)
		return
	}

	// log authentication
	app.logRequest("success logging", fmt.Sprintf("User with email %s successully logged in", user.Email))

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name, data string) {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	byteArray, _ := json.Marshal(entry)

	request, _ := http.NewRequest("PUT", "http://logger-service/log", bytes.NewBuffer(byteArray))
	request.Header.Set(headers.ContentType, "application/json")

	defer request.Body.Close()

	client := &http.Client{}

	resp, err := client.Do(request)

	if err != nil || resp.StatusCode != http.StatusAccepted {
		log.Panic("Can't send log... PS Huinya Vasya datay vse po novoy")
	}
}
