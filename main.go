package main

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	lgts := *newlgts()
	slackToken := os.Getenv("SLACK_TOKEN")
	go runrtm(slackToken)

	router.GET("/apps", lgts.getApps)
	router.POST("/apps", lgts.registerApp)

	router.GET("/messages", lgts.getMessages)
	router.POST("/messages", lgts.registerMessage)

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "localhost:8080"
	}

	log.Printf("Starting lgts server on %s\n", serverURL)
	log.Fatal(http.ListenAndServe(serverURL, router))
}
