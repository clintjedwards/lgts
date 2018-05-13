package main

import (
	"fmt"
	"math/rand"
	"time"
)

type trackedMessage struct {
	ID          string   `json:"id"`           //The unique message id that is used to identify the slack message
	Submitted   int64    `json:"submitted"`    //Time the request was submitted in epoch
	CallbackURL string   `json:"callback_url"` //URL to send event stream of emoji usage
	ValidEmojis []string `json:"valid_emojis"` //List of emojis to alert on
	AuthToken   string   `json:"auth_token"`   //Auth token given by app to auth on callback
	Expire      int      `json:"expire"`       //Length of time messages can be tracked. Limited to 24h
}

type messageEvent struct {
	ID        string `json:"id"`         //The unique message ID
	EmojiUsed string `json:"emoji_used"` //EmojiUsed in event
	AuthToken string `json:"auth_token"` //Pre-shared auth token given
	SlackUser string `json:"slack_user"` //Slack user who used the emoji
}

func (app *app) getMessage(messageID string) (trackedMessage, error) {

	if _, ok := app.messages[messageID]; !ok {
		return trackedMessage{}, errMessageNotFound
	}

	message := app.messages[messageID]

	return message, nil
}

func (app *app) createMessage(callbackURL, authToken string, validEmojs []string) trackedMessage {

	messageID := app.generateNewMessageID()

	newMessage := &trackedMessage{
		ID:          messageID,
		Submitted:   time.Now().Unix(),
		CallbackURL: callbackURL,
		AuthToken:   authToken,
		ValidEmojis: validEmojs,
	}

	app.messages[messageID] = *newMessage

	return *newMessage
}

func (app *app) deleteMessage(messageID string) error {
	if _, ok := app.messages[messageID]; !ok {
		return errMessageNotFound
	}

	delete(app.messages, messageID)

	return nil
}

func generateRandomString(length int) string {

	rand.Seed(time.Now().UTC().UnixNano())
	token := make([]byte, length)
	rand.Read(token)

	return fmt.Sprintf("%x", token)
}

func (app *app) generateNewMessageID() string {

	var messageID string

	for {
		messageID := generateRandomString(10)

		if _, exists := app.messages[messageID]; !exists {
			break
		}
	}

	return messageID
}
