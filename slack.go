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

func generateMessageAttachment(api *slack.Client, messageTimestamp string, username, channel string) slack.Attachment {

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
		Value: fmt.Sprintf("Approved by %s", username),
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

	switch {
	case lgts.isApprovalEmoji(emojiUsed):
		//set some stuff here for each case
	case lgts.isRejectionEmoji(emojiUsed):

	default:
		return
	}

	callbackInfo := getMessageCallbackInfo(api, messageTimestamp, channel)

	if _, ok := callbackInfo["message_token"]; !ok {
		log.Printf("message_token parameter missing from callback id string")
		return
	}
	if _, ok := callbackInfo["app_id"]; !ok {
		log.Printf("app_id parameter missing from callback id string")
		return
	}

	hashedPair := getSHA1(callbackInfo["app_id"].(string), callbackInfo["message_token"].(string))
	appID := lgts.Messages[hashedPair].AppID
	app := lgts.Apps[appID]
	fmt.Println(app)

	userInfo, err := getUser(api, userID)
	if err != nil {
		log.Fatalf("Cannot find slack user: %v", err)
	}

	attachment := generateMessageAttachment(api, messageTimestamp, userInfo.fullName, channel)
	updateMessage(api, attachment, messageTimestamp, channel)

	if lgts.isAuthorizedUser(callbackInfo["app_id"].(string), userInfo.email) {
		log.Println(userInfo.email)
	}

}

// func processApproveOrReject(lgts *lgts, emojiUsed string) {

// }

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
