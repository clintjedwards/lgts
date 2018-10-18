package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/clintjedwards/snark/config"
	"github.com/clintjedwards/snark/helpers/httputil"
	"github.com/nlopes/slack"
)

//We define two slackClients because each token has different permissions
// that do not and cannot overlap
type app struct {
	config         *config.Config
	slackAppClient *slack.Client
	slackBotClient *slack.Client
	messages       map[string]*trackedMessage //map with messageID as the key
}

func newApp() *app {

	config, err := config.FromEnv()
	if err != nil {
		log.Fatal(err)
	}

	slackAppClient := slack.New(config.Slack.AppToken)
	slackBotClient := slack.New(config.Slack.BotToken)

	return &app{
		config:         config,
		slackAppClient: slackAppClient,
		slackBotClient: slackBotClient,
		messages:       make(map[string]*trackedMessage),
	}
}

func (app *app) isValidEmoji(messageID, emoji string) bool {

	for _, validEmoji := range app.messages[messageID].ValidEmojis {
		if validEmoji == emoji {
			return true
		}
	}
	return false
}

func (app *app) sendEvent(messageEvent messageEvent) error {

	trackedMessage, err := app.getMessage(messageEvent.ID)
	if err != nil {
		return err
	}

	jsonString, _ := json.Marshal(&messageEvent)

	headers := map[string]string{
		"content-type":     "application/json",
		"snark-auth-token": trackedMessage.AuthToken,
	}

	response, err := httputil.SendHTTPPOSTRequest(trackedMessage.CallbackURL, headers, jsonString, app.config.Debug)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		err := fmt.Errorf("callback URL did not return correct status code; recieved %d", response.StatusCode)
		return err
	}

	return nil
}

func (app *app) getVersionHandler(w http.ResponseWriter, req *http.Request) {

	response := struct {
		Version string `json:"Version"`
	}{Version: "0.1.0"}

	sendResponse(w, http.StatusOK, response, false)
	return
}
