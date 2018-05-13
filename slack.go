package main

import (
	"fmt"
	"log"
	"time"

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

//processSlackMessage grabs all message data to be evaluated. If message data doesn't have required string
// we return error
func (app *app) processSlackMessage(event *slack.ReactionAddedEvent) error {

	messageID, err := app.getSlackMessageID(event.Item.Timestamp, event.Item.Channel)
	if err != nil {
		return err
	}

	trackedMessage, err := app.getMessage(messageID)
	if err != nil {
		return err
	}

	userInfo, err := app.getSlackUser(event.User)
	if err != nil {
		err := fmt.Errorf("Cannot find slack user: %s", err)
		return err
	}

	valid := app.isValidEmoji(messageID, event.Reaction)
	if !valid {
		err := fmt.Errorf("Emoji %s is not a valid emoji for messageID %s", event.Reaction, messageID)
		return err
	}

	newEventMessage := messageEvent{
		ID:             trackedMessage.ID,
		Submitted:      time.Now().Unix(),
		EmojiUsed:      event.Reaction,
		SlackUserName:  userInfo.fullName,
		SlackUserEmail: userInfo.email,
	}

	trackedMessage.MessageEvents = append(trackedMessage.MessageEvents, newEventMessage)

	if trackedMessage.CallbackURL != "" {
		err = app.sendEvent(newEventMessage)
		if err != nil {
			return err
		}
	}

	log.Printf("emoji %s was applied to message id %s by slack user %s", event.Reaction, messageID, userInfo.fullName)

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
				err := fmt.Errorf("[Slack] %v", err)
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
