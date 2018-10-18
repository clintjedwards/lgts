package main

import (
	"fmt"
	"log"
	"math/rand"
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
		err := fmt.Errorf("cannot find slack user: %s", err)
		return err
	}

	valid := app.isValidEmoji(messageID, event.Reaction)
	if !valid {
		err := fmt.Errorf("emoji %s is not a valid emoji for message id %s", event.Reaction, messageID)
		return err
	}

	newEventMessage := messageEvent{
		ID:             trackedMessage.ID,
		Submitted:      time.Now().Unix(),
		EmojiUsed:      event.Reaction,
		SlackUserName:  userInfo.fullName,
		SlackUserEmail: userInfo.email,
	}

	trackedMessage.MessageEvents = append(trackedMessage.MessageEvents, &newEventMessage)

	if trackedMessage.CallbackURL != "" {
		err = app.sendEvent(newEventMessage)
		if err != nil {
			return err
		}
	}

	log.Printf("emoji %s was applied to message id %s by slack user %s", event.Reaction, messageID, userInfo.fullName)

	return nil
}

// mockrtm generates fake emoji responses at random. Used primarily for testing
func (app *app) mockrtm() {

	log.Println("dev mode enabled")

	reactionPool := []string{":shrug:", ":+1:", ":-1:"}

	for {
		for message, info := range app.messages {
			rand.Seed(time.Now().UnixNano())
			emoji := reactionPool[rand.Intn(len(reactionPool))]

			info.MessageEvents = append(info.MessageEvents, &messageEvent{
				Submitted:      time.Now().Unix(),
				EmojiUsed:      emoji,
				SlackUserName:  "barack.obama",
				SlackUserEmail: "barack.obama@usa.com",
			})

			log.Printf("dev mode: added %s emoji for message %s\n", emoji, message)
		}
		time.Sleep(time.Second * 2)
	}
}

// runrtm runs slack's real time event stream and listens for reaction events
func (app *app) runrtm() {

	if app.config.Dev {
		app.mockrtm()
		return
	}

	rtm := app.slackBotClient.NewRTM()
	go rtm.ManageConnection()

	log.Println("slack: starting slack event reader")

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ReactionAddedEvent:
			err := app.processSlackMessage(ev)
			if err != nil {
				err := fmt.Errorf("[slack] %v", err)
				log.Println(err)
			}

		case *slack.RTMError:
			log.Printf("error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			log.Printf("invalid credentials")
			return

		default:
			// Ignore other events
		}
	}
}
