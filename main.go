package main

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

var slackToken = os.Getenv("SLACK_TOKEN")
var serverURL = os.Getenv("SERVER_URL")
var stateFilePath = os.Getenv("STATE_FILE_PATH")

func init() {

	if slackToken == "" {
		log.Fatal("$SLACK_TOKEN not set")
	}

	if serverURL == "" {
		serverURL = "localhost:8080"
	}

	if stateFilePath == "" {
		stateFilePath = "./state"
	}

}

func main() {
	router := httprouter.New()

	lgts := *newlgts(stateFilePath)
	lgts.loadState()
	go runrtm(&lgts, slackToken)

	router.GET("/apps", lgts.getApps)
	router.POST("/apps", lgts.registerApp)

	router.GET("/messages", lgts.getMessages)
	router.POST("/messages", lgts.registerMessage)

	log.Printf("Starting lgts server on %s\n", serverURL)
	log.Fatal(http.ListenAndServe(serverURL, router))
}
