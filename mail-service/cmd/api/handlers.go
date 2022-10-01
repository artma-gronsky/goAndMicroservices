package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		log.Println(err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = app.Mailer.SendSMTPMessage(msg)

	if err != nil {
		js, _ := json.MarshalIndent(msg, "", "\t")
		log.Println(bytes.NewBuffer(js))
		log.Println(err)
		app.errorJSON(w, errors.New("Problem with sending mail"), http.StatusBadRequest)
		return
	}

	app.writeJson(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: "Mail was sent",
	})
}
