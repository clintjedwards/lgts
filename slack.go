package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nlopes/slack"
)

type user struct {
	fullName string
	email    string
}

func getMessageCallbackInfo(api *slack.Client, messageTimestamp, channel string) (map[string]interface{}, error) {
	history, err := api.GetChannelHistory(channel, slack.HistoryParameters{
		Count:     1,
		Inclusive: true,
		Latest:    messageTimestamp,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	messageInfo := history.Messages[0].Msg.Attachments[0].CallbackID

	var callbackInfo map[string]interface{}
	err = json.Unmarshal([]byte(messageInfo), &callbackInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if _, ok := callbackInfo["message_id"]; !ok {
		err := fmt.Errorf("message_id parameter missing from callback id string")
		log.Println(err)
		return nil, err
	}
	if _, ok := callbackInfo["app_name"]; !ok {
		err := fmt.Errorf("app_name parameter missing from callback id string")
		log.Println(err)
		return nil, err
	}

	callbackInfo["channel"] = channel
	callbackInfo["message_timestamp"] = messageTimestamp

	return callbackInfo, nil

}

func getUser(api *slack.Client, userID string) (*user, error) {
	userInfo, err := api.GetUserInfo(userID)
	if err != nil {
		log.Printf("%s\n", err)
		return nil, err
	}

	return &user{userInfo.Profile.RealName, userInfo.Profile.Email}, nil

}

func processDecision(api *slack.Client, lgts *lgts, event *slack.ReactionAddedEvent) {

	userID := event.User
	channel := event.Item.Channel
	messageTimestamp := event.Item.Timestamp
	emojiUsed := event.Reaction

	//Check if the emoji used is one of the predefined emojis
	var isApproved bool
	var decisionVerb string

	switch {
	case lgts.isApprovalEmoji(emojiUsed):
		isApproved = true
		decisionVerb = "approved"
	case lgts.isRejectionEmoji(emojiUsed):
		isApproved = false
		decisionVerb = "rejected"
	default:
		return
	}

	//Get information about the message
	callbackInfo, err := getMessageCallbackInfo(api, messageTimestamp, channel)
	if err != nil {
		log.Printf("Cannot not successfully get callback information. Skipping message. %v", err)
		return
	}

	//Use information to recieve proper app/message object and check for existence
	message, present := lgts.Messages[callbackInfo["message_id"].(string)]
	if !present {
		log.Println("Message processed but not found in queue")
		return
	}

	app, present := lgts.Apps[message.AppName]
	if !present {
		log.Println("Message processed but app not registered")
		return
	}

	//Check if user who used the emoji is part of approved list
	userInfo, err := getUser(api, userID)
	if err != nil {
		log.Printf("Cannot find slack user: %v", err)
		return
	}

	if !app.isAuthorizedUser(userInfo.email) {
		log.Println("User not authorized to approve")
		return
	}

	err = app.sendMessageApproval(callbackInfo, userInfo.email, isApproved)
	if err != nil {
		log.Printf("Couldn't send proper request: %v", err)
		return
	}

	log.Printf("Message ID %s was %s by slack user %s", message.ID, decisionVerb, userInfo.fullName)

	delete(lgts.Messages, message.ID)

}

func runrtm(lgts *lgts, slackToken string, debug bool) {

	api := slack.New(slackToken)
	api.SetDebug(debug)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	log.Println("Starting slack event reader")

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ReactionAddedEvent:
			processDecision(api, lgts, ev)

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
