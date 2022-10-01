package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Config struct {
	Mailer Mail
}

const webPort = 80

func main() {
	app := Config{
		Mailer: createMail(),
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", webPort),
		Handler: app.routes(),
	}

	err := httpServer.ListenAndServe()

	if err != nil {
		log.Panic("Can't start server:" + err.Error())
	}
}

func createMail() Mail {
	var m Mail

	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))

	m.Domain = os.Getenv("MAIL_DOMAIN")
	m.Host = os.Getenv("MAIL_HOST")
	m.Port = port
	m.Username = os.Getenv("MAIL_USERNAME")
	m.Password = os.Getenv("MAIL_PASSWORD")
	m.Encryption = os.Getenv("MAIL_ENCRYPTION")
	m.FromName = os.Getenv("MAIL_FROM_NAME")
	m.FromAddress = os.Getenv("MAIL_FROM_ADDRESS")

	return m
}
