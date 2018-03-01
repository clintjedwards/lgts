package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type message struct {
	Created time.Time `json:"created"`
}

func (app *app) getMessage(serviceName, messageID string) (message, error) {

	if _, ok := app.messages[serviceName]; !ok {
		return message{}, errMessageNotFound
	}

	if _, ok := app.messages[serviceName][messageID]; !ok {
		return message{}, errMessageNotFound
	}

	message := app.messages[serviceName][messageID]

	return message, nil
}

func (app *app) getMessages(serviceName string) (map[string]message, error) {

	if _, ok := app.messages[serviceName]; !ok {
		return map[string]message{}, nil
	}

	return app.messages[serviceName], nil
}

func (app *app) createMessage(serviceName string) (messageID string, err error) {

	messageID = generateRandomString(15)

	newMessage := &message{
		Created: time.Now(),
	}

	if _, ok := app.messages[serviceName][messageID]; ok {
		return "", errMessageExists
	}

	if app.messages[serviceName] == nil {
		app.messages[serviceName] = make(map[string]message)
	}

	app.messages[serviceName][messageID] = *newMessage

	return messageID, nil
}

func (app *app) deleteMessage(serviceName, messageID string) error {
	if _, ok := app.messages[serviceName][messageID]; !ok {
		return errMessageNotFound
	}

	delete(app.messages[serviceName], messageID)

	return nil
}

func generateRandomString(length int) string {

	rand.Seed(time.Now().UTC().UnixNano())
	token := make([]byte, length)
	rand.Read(token)

	return fmt.Sprintf("%x", token)
}

func obfuscateString(str string) string {
	const maskPercentage = .30
	strLen := len(str)

	showLength := float64(strLen) * maskPercentage
	showLengthRounded := int(showLength)

	obfuscatedString := strings.Repeat("x", strLen-showLengthRounded) + str[len(str)-3:]

	return obfuscatedString
}
