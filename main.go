// lgts (looks good to slack) listens to slack event stream
// and then reports back when certain emojis are used by certain
// users on predetermined messages

// Useful for instances where public callbacks are impossible

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

var slackToken = os.Getenv("SLACK_TOKEN")
var serverURL = os.Getenv("SERVER_URL")
var stateFilePath = os.Getenv("STATE_FILE_PATH") // Full path to save state of registered applications
var debug bool
var debugString = os.Getenv("DEBUG")

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

	if debugString == "" {
		debug = false
	} else {
		debug, _ = strconv.ParseBool(debugString)
	}
}

func main() {
	router := httprouter.New()

	lgts := *newlgts(stateFilePath)
	lgts.loadState()
	go runrtm(&lgts, slackToken, debug)

	router.GET("/apps", lgts.getApps)
	router.POST("/apps", lgts.registerApp)

	router.GET("/messages", lgts.getMessages)
	router.POST("/messages", lgts.registerMessage)

	log.Printf("Starting lgts server on %s\n", serverURL)
	log.Fatal(http.ListenAndServe(serverURL, router))
}
