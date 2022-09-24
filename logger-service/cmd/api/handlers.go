package main

import (
	"errors"
	"log"
	"logger-service/data"
	"net/http"
)

type writeLogRequest struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var request writeLogRequest

	if err := app.readJSON(w, r, &request); err != nil {
		err = errors.New("Error of reading request body: " + err.Error())
		log.Println(err.Error())
		app.errorJSON(w, err)
		return
	}

	err := app.Models.LogEntry.Insert(data.LogEntry{
		Data: request.Data,
		Name: request.Name,
	})

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "logged",
	}

	app.writeJson(w, http.StatusAccepted, resp)
}
