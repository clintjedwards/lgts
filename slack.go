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

func generateMessageAttachment(api *slack.Client, messageTimestamp, username, channel, decisionVerb string) slack.Attachment {

	history, err := api.GetChannelHistory(channel, slack.HistoryParameters{
		Count:     1,
		Inclusive: true,
		Latest:    messageTimestamp,
	})
	if err != nil {
		log.Println(err)
	}

	messageAttachment := history.Messages[0].Msg.Attachments[0]
	messageAttachment.Fields[0] = slack.AttachmentField{
		Value: fmt.Sprintf("%s by %s", decisionVerb, username),
		Short: false,
	}

	return messageAttachment
}

func updateMessage(api *slack.Client, attachment slack.Attachment, messageTimestamp, channel string) {
	//We use SendMessage here instead of UpdateMessage because the latter does not support updating attachments
	// https://github.com/nlopes/slack/pull/121
	api.SendMessage(channel, slack.MsgOptionUpdate(messageTimestamp), slack.MsgOptionAttachments(attachment))
}

func getMessageCallbackInfo(api *slack.Client, messageTimestamp, channel string) map[string]interface{} {
	history, err := api.GetChannelHistory(channel, slack.HistoryParameters{
		Count:     1,
		Inclusive: true,
		Latest:    messageTimestamp,
	})
	if err != nil {
		log.Println(err)
	}

	messageInfo := history.Messages[0].Msg.Attachments[0].CallbackID

	var callbackInfo map[string]interface{}
	json.Unmarshal([]byte(messageInfo), &callbackInfo)

	return callbackInfo

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
	callbackInfo := getMessageCallbackInfo(api, messageTimestamp, channel)

	if _, ok := callbackInfo["message_id"]; !ok {
		log.Printf("message_id parameter missing from callback id string")
		return
	}
	if _, ok := callbackInfo["app_name"]; !ok {
		log.Printf("app_name parameter missing from callback id string")
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

	//User, emoji, and messageid check out
	// update slack message and send callback url proper message

	err = app.sendMessageApproval(callbackInfo, userInfo.email, isApproved)
	if err != nil {
		//return
		log.Printf("Couldn't send proper request: %v", err)
	}

	attachment := generateMessageAttachment(api, messageTimestamp, userInfo.fullName, channel, decisionVerb)
	updateMessage(api, attachment, messageTimestamp, channel)

	log.Printf("Message ID %s was %s by slack user %s", message.ID, decisionVerb, userInfo.fullName)

	delete(lgts.Messages, message.ID)

}

func runrtm(lgts *lgts, slackToken string) {

	api := slack.New(slackToken)
	//api.SetDebug(true)

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
