package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type lgts struct {
	Apps           map[string]app     `json:"apps"`     //registered applications
	Messages       map[string]message `json:"messages"` // messages to process, removed once sent back to client
	ApprovalEmojis []string           `json:"approval_emojis"`
	RejectEmojis   []string           `json:"reject_emojis"`
}

func newlgts() *lgts {

	return &lgts{
		Apps:           make(map[string]app),
		Messages:       make(map[string]message),
		ApprovalEmojis: make([]string, 0),
		RejectEmojis:   make([]string, 0),
	}

}

func (lgts *lgts) getApps(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lgts.Apps)
}

func (lgts *lgts) registerApp(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	newapp := newApp()

	var appinfo app

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&appinfo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(struct {
			Ok      bool   `json:"ok"`
			Message string `json:"message"`
			Error   string `json:"error"`
		}{false, "could not decode json body", fmt.Sprintf("%v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if appinfo.Name == "" || appinfo.CallbackURL == "" || len(appinfo.AuthorizedApprovers) == 0 {

		w.WriteHeader(400)
		json.NewEncoder(w).Encode(struct {
			Ok      bool   `json:"ok"`
			Message string `json:"message"`
		}{
			false,
			"Invalid Parameters: You must supply name, callbackurl, authorizedapprovers as params",
		})

	} else {

		newapp.Name = strings.TrimSpace(appinfo.Name)
		newapp.CallbackURL = strings.TrimSpace(appinfo.CallbackURL)
		for i := range appinfo.AuthorizedApprovers {
			newapp.AuthorizedApprovers = append(newapp.AuthorizedApprovers, strings.TrimSpace(appinfo.AuthorizedApprovers[i]))
		}

		lgts.Apps[newapp.ID] = *newapp

		json.NewEncoder(w).Encode(struct {
			ID   string `json:"app_id"`
			Name string `json:"app_name"`
		}{newapp.ID, newapp.Name})

		log.Printf("Application %s:%s registered", newapp.Name, newapp.ID)
	}

}

func (lgts *lgts) unregisterApp(appID string) {
	delete(lgts.Apps, appID)
}

func (lgts *lgts) registerMessage(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	var messageInfo struct {
		ID    string `json:"id"`
		AppID string `json:"app_id"`
		Token string `json:"message_token"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&messageInfo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(struct {
			Ok      bool   `json:"ok"`
			Message string `json:"message"`
			Error   string `json:"error"`
		}{false, "could not decode json body", fmt.Sprintf("%v", err)})
		return
	}

	if messageInfo.AppID == "" || messageInfo.Token == "" {

		w.WriteHeader(400)
		json.NewEncoder(w).Encode(struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}{"Invalid Parameters", "You must supply both app_id and message_token params"})
		return

	} else if !lgts.isApplicationRegistered(messageInfo.AppID) {

		log.Println("Application ID not registered")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(struct {
			Ok      bool   `json:"ok"`
			Message string `json:"message"`
		}{false, "Application ID not registered"})
		return

	} else {
		newMessage := *newMessage()
		newMessage.AppID = messageInfo.AppID
		newMessage.token = messageInfo.Token

		lgts.Messages[newMessage.ID] = newMessage

		json.NewEncoder(w).Encode(newMessage)

		log.Printf("Message %s registered from application: %s", newMessage.ID, newMessage.AppID)
	}
}

func (lgts *lgts) isApplicationRegistered(appID string) bool {

	for apps := range lgts.Apps {
		if apps == appID {
			return true
		}
	}
	return false
}

func getSHA1(args ...string) string {
	unhashedString := strings.Join(args, "")

	h := sha1.New()
	h.Write([]byte(unhashedString))
	hashedString := h.Sum(nil)

	return fmt.Sprintf("%x", hashedString)
}

func (lgts *lgts) getMessages(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lgts.Messages)
}

func (lgts *lgts) isAuthorizedUser(appID, email string) bool {

	app := lgts.Apps[appID]

	for _, approvedEmail := range app.AuthorizedApprovers {
		if approvedEmail == email {
			return true
		}
	}
	return false

}

func (lgts *lgts) isApprovalEmoji(emoji string) bool {

	for _, approvalEmoji := range lgts.ApprovalEmojis {
		if approvalEmoji == emoji {
			return true
		}
	}
	return false

}

func (lgts *lgts) isRejectionEmoji(emoji string) bool {

	for _, rejectEmoji := range lgts.RejectEmojis {
		if rejectEmoji == emoji {
			return true
		}
	}
	return false

}
