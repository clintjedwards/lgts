package main

import (
	"fmt"
	"log"

	"github.com/nlopes/slack"
)

type user struct {
	fullName string
	email    string
}

func (app *app) getSlackMessageID(messageTimestamp, channel string) (messageID string, err error) {
	history, err := app.slackAppClient.GetChannelHistory(channel, slack.HistoryParameters{
		Count:     1,
		Inclusive: true,
		Latest:    messageTimestamp,
	})
	if err != nil {
		log.Printf("slack: %s", err)
		return "", err
	}

	if len(history.Messages[0].Msg.Attachments) > 0 {
		messageID = history.Messages[0].Msg.Attachments[0].CallbackID
	} else {
		err := fmt.Errorf("message processed; no attachment found")
		return "", err
	}

	if messageID == "" {
		err := fmt.Errorf("missing messageID in attachment body/callbackID param")
		return "", err
	}

	return messageID, nil

}

func (app *app) getSlackUser(userID string) (*user, error) {
	userInfo, err := app.slackAppClient.GetUserInfo(userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &user{userInfo.Profile.RealName, userInfo.Profile.Email}, nil
}

//Used to search service for the service with a specific messageID
// might be useful to consider a global map to store this in instead to speed things up
func (app *app) getServicebyMessageID(messageID string) (*service, error) {
	services, err := app.getServices()
	if err != nil {
		return &service{}, err
	}

	referencedService := &service{}
	for _, service := range services {
		if (app.messages[service.Name][messageID]) != (message{}) {
			referencedService = service
			break
		}
	}

	if referencedService.Name == "" {
		err := fmt.Errorf("Cannot find service with messageID %s", obfuscateString(messageID))
		return &service{}, err
	}

	return referencedService, err
}

//parseMessage grabs all message data to be evaluation. If message data doesn't have required string
// we return error
func (app *app) processSlackMessage(event *slack.ReactionAddedEvent) error {

	messageID, err := app.getSlackMessageID(event.Item.Timestamp, event.Item.Channel)
	if err != nil {
		return err
	}

	service, err := app.getServicebyMessageID(messageID)
	if err != nil {
		return err
	}

	//Check if user who used the emoji is part of approved list
	userInfo, err := app.getSlackUser(event.User)
	if err != nil {
		err := fmt.Errorf("Cannot find slack user: %s", err)
		return err
	}

	if !service.isAuthorizedSlacker(userInfo.email) {
		err := fmt.Errorf("User %s not authorized to approve messageID %s", userInfo.email, obfuscateString(messageID))
		return err
	}

	err = service.sendCallbackMessage(messageID, userInfo.email, event.Reaction)
	if err != nil {
		err := fmt.Errorf("Couldn't send request to callback URL %s for service %s: %s", service.CallbackURL, service.Name, err)
		return err
	}

	log.Printf("emoji %s was applied to message id %s by slack user %s; removing message from queue", event.Reaction, messageID, userInfo.fullName)

	err = app.deleteMessage(service.Name, messageID)
	if err != nil {
		return err
	}

	return nil
}

// runrtm runs slack's real time event stream and listens for reaction events
func (app *app) runrtm() {

	rtm := app.slackBotClient.NewRTM()
	go rtm.ManageConnection()

	log.Println("Slack: Starting slack event reader")

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ReactionAddedEvent:
			err := app.processSlackMessage(ev)
			if err != nil {
				err := fmt.Errorf("Slack: %v", err)
				log.Println(err)
				return
			}

		case *slack.RTMError:
			log.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			log.Printf("Invalid credentials")
			return

		default:
			// Ignore other events..
		}
	}
}
