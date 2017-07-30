package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nlopes/slack"
)

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

func isAuthorizedUser() {

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

func getUser(api *slack.Client, userID string) [2]string {
	user, err := api.GetUserInfo(userID)
	if err != nil {
		log.Printf("%s\n", err)
		return [2]string{}
	}

	return [2]string{user.Profile.RealName, user.Profile.Email}
}

func runrtm(slackToken string) {

	api := slack.New(slackToken)
	//api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	log.Println("Starting slack event reader")

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ReactionAddedEvent:
			userInfo := getUser(api, ev.User)
			messageTimestamp := ev.Item.Timestamp
			attachment := generateMessageAttachment(api, messageTimestamp, userInfo[0], ev.Item.Channel)
			updateMessage(api, attachment, messageTimestamp, ev.Item.Channel)
			callbackInfo := getMessageCallbackInfo(api, messageTimestamp, ev.Item.Channel)
			//callbackInfo["lgts_token"]
			//callbackInfo[""]
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
